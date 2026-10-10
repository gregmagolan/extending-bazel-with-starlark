// Package device holds the device domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package device

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// DeviceRecordSchema is the stored shape of a DeviceRecord.
var DeviceRecordSchema = schema.New("device.devicerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceRecord is one device record.
type DeviceRecord struct {
	ID        string
	Comment   string
	Reference string
	Label     string
	Priority  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeviceRecord) Key() string { return "device/devicerecord/" + x.ID }

// Fields encodes x for storage.
func (x DeviceRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"label":     x.Label,
		"priority":  strconv.Itoa(x.Priority),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceRecord decodes a stored record.
func ParseDeviceRecord(f codec.Fields) (DeviceRecord, error) {
	var x DeviceRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicerecord: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("device: devicerecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("device: devicerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceRecord) String() string {
	return fmt.Sprintf("DeviceRecord(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceRecord) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceRecord) Normalize() DeviceRecord {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDeviceRecords orders by priority descending, then by ID.
func SortDeviceRecords(xs []DeviceRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceRecords drops the archived entries, preserving order.
func ActiveDeviceRecords(xs []DeviceRecord) []DeviceRecord {
	out := make([]DeviceRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceRecords adds up the priority of every active entry.
func SumDeviceRecords(xs []DeviceRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeviceRecords keys the entries by ID, last write winning.
func IndexDeviceRecords(xs []DeviceRecord) map[string]DeviceRecord {
	out := make(map[string]DeviceRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceRecord builds a valid DeviceRecord, for tests and for the seed data the cron
// binary writes.
func SampleDeviceRecord(id string, at time.Time) DeviceRecord {
	return DeviceRecord{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Label:     "seeded by device.SampleDeviceRecord",
		Priority:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeviceEntrySchema is the stored shape of a DeviceEntry.
var DeviceEntrySchema = schema.New("device.deviceentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceEntry is one device entry.
type DeviceEntry struct {
	ID       string
	Source   string
	Channel  string
	Region   string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceEntry) Key() string { return "device/deviceentry/" + x.ID }

// Fields encodes x for storage.
func (x DeviceEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"region":   x.Region,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceEntry decodes a stored record.
func ParseDeviceEntry(f codec.Fields) (DeviceEntry, error) {
	var x DeviceEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: deviceentry: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("device: deviceentry %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("device: deviceentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: deviceentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: deviceentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceEntry) String() string {
	return fmt.Sprintf("DeviceEntry(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceEntry) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceEntry) Normalize() DeviceEntry {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeviceEntrys orders by version descending, then by ID.
func SortDeviceEntrys(xs []DeviceEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceEntrys drops the archived entries, preserving order.
func ActiveDeviceEntrys(xs []DeviceEntry) []DeviceEntry {
	out := make([]DeviceEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceEntrys adds up the version of every active entry.
func SumDeviceEntrys(xs []DeviceEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeviceEntrys keys the entries by ID, last write winning.
func IndexDeviceEntrys(xs []DeviceEntry) map[string]DeviceEntry {
	out := make(map[string]DeviceEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceEntry builds a valid DeviceEntry, for tests and for the seed data the cron
// binary writes.
func SampleDeviceEntry(id string, at time.Time) DeviceEntry {
	return DeviceEntry{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by device.SampleDeviceEntry",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceItemSchema is the stored shape of a DeviceItem.
var DeviceItemSchema = schema.New("device.deviceitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceItem is one device item.
type DeviceItem struct {
	ID        string
	Reference string
	Label     string
	Channel   string
	Retained  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeviceItem) Key() string { return "device/deviceitem/" + x.ID }

// Fields encodes x for storage.
func (x DeviceItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"channel":   x.Channel,
		"retained":  strconv.Itoa(x.Retained),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceItem decodes a stored record.
func ParseDeviceItem(f codec.Fields) (DeviceItem, error) {
	var x DeviceItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: deviceitem: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("device: deviceitem %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: deviceitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: deviceitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: deviceitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceItem) String() string {
	return fmt.Sprintf("DeviceItem(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceItem) Normalize() DeviceItem {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDeviceItems orders by retained descending, then by ID.
func SortDeviceItems(xs []DeviceItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceItems drops the archived entries, preserving order.
func ActiveDeviceItems(xs []DeviceItem) []DeviceItem {
	out := make([]DeviceItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceItems adds up the retained of every active entry.
func SumDeviceItems(xs []DeviceItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDeviceItems keys the entries by ID, last write winning.
func IndexDeviceItems(xs []DeviceItem) map[string]DeviceItem {
	out := make(map[string]DeviceItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceItem builds a valid DeviceItem, for tests and for the seed data the cron
// binary writes.
func SampleDeviceItem(id string, at time.Time) DeviceItem {
	return DeviceItem{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Channel:   "seeded by device.SampleDeviceItem",
		Retained:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeviceDraftSchema is the stored shape of a DeviceDraft.
var DeviceDraftSchema = schema.New("device.devicedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceDraft is one device draft.
type DeviceDraft struct {
	ID       string
	Name     string
	Owner    string
	Label    string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceDraft) Key() string { return "device/devicedraft/" + x.ID }

// Fields encodes x for storage.
func (x DeviceDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceDraft decodes a stored record.
func ParseDeviceDraft(f codec.Fields) (DeviceDraft, error) {
	var x DeviceDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicedraft: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("device: devicedraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("device: devicedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceDraft) String() string {
	return fmt.Sprintf("DeviceDraft(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceDraft) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceDraft) Normalize() DeviceDraft {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeviceDrafts orders by retained descending, then by ID.
func SortDeviceDrafts(xs []DeviceDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceDrafts drops the archived entries, preserving order.
func ActiveDeviceDrafts(xs []DeviceDraft) []DeviceDraft {
	out := make([]DeviceDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceDrafts adds up the retained of every active entry.
func SumDeviceDrafts(xs []DeviceDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDeviceDrafts keys the entries by ID, last write winning.
func IndexDeviceDrafts(xs []DeviceDraft) map[string]DeviceDraft {
	out := make(map[string]DeviceDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceDraft builds a valid DeviceDraft, for tests and for the seed data the cron
// binary writes.
func SampleDeviceDraft(id string, at time.Time) DeviceDraft {
	return DeviceDraft{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by device.SampleDeviceDraft",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceRevisionSchema is the stored shape of a DeviceRevision.
var DeviceRevisionSchema = schema.New("device.devicerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceRevision is one device revision.
type DeviceRevision struct {
	ID       string
	Channel  string
	Owner    string
	Label    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceRevision) Key() string { return "device/devicerevision/" + x.ID }

// Fields encodes x for storage.
func (x DeviceRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceRevision decodes a stored record.
func ParseDeviceRevision(f codec.Fields) (DeviceRevision, error) {
	var x DeviceRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicerevision: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: devicerevision %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("device: devicerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceRevision) String() string {
	return fmt.Sprintf("DeviceRevision(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceRevision) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceRevision) Normalize() DeviceRevision {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDeviceRevisions orders by capacity descending, then by ID.
func SortDeviceRevisions(xs []DeviceRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceRevisions drops the archived entries, preserving order.
func ActiveDeviceRevisions(xs []DeviceRevision) []DeviceRevision {
	out := make([]DeviceRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceRevisions adds up the capacity of every active entry.
func SumDeviceRevisions(xs []DeviceRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDeviceRevisions keys the entries by ID, last write winning.
func IndexDeviceRevisions(xs []DeviceRevision) map[string]DeviceRevision {
	out := make(map[string]DeviceRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceRevision builds a valid DeviceRevision, for tests and for the seed data the cron
// binary writes.
func SampleDeviceRevision(id string, at time.Time) DeviceRevision {
	return DeviceRevision{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by device.SampleDeviceRevision",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceSnapshotSchema is the stored shape of a DeviceSnapshot.
var DeviceSnapshotSchema = schema.New("device.devicesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceSnapshot is one device snapshot.
type DeviceSnapshot struct {
	ID       string
	Comment  string
	Region   string
	Channel  string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceSnapshot) Key() string { return "device/devicesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x DeviceSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceSnapshot decodes a stored record.
func ParseDeviceSnapshot(f codec.Fields) (DeviceSnapshot, error) {
	var x DeviceSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicesnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("device: devicesnapshot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("device: devicesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceSnapshot) String() string {
	return fmt.Sprintf("DeviceSnapshot(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceSnapshot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceSnapshot) Normalize() DeviceSnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeviceSnapshots orders by count descending, then by ID.
func SortDeviceSnapshots(xs []DeviceSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceSnapshots drops the archived entries, preserving order.
func ActiveDeviceSnapshots(xs []DeviceSnapshot) []DeviceSnapshot {
	out := make([]DeviceSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceSnapshots adds up the count of every active entry.
func SumDeviceSnapshots(xs []DeviceSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDeviceSnapshots keys the entries by ID, last write winning.
func IndexDeviceSnapshots(xs []DeviceSnapshot) map[string]DeviceSnapshot {
	out := make(map[string]DeviceSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceSnapshot builds a valid DeviceSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleDeviceSnapshot(id string, at time.Time) DeviceSnapshot {
	return DeviceSnapshot{
		ID:       id,
		Comment:  "comment-" + id,
		Region:   "region-" + id,
		Channel:  "seeded by device.SampleDeviceSnapshot",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceBatchSchema is the stored shape of a DeviceBatch.
var DeviceBatchSchema = schema.New("device.devicebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceBatch is one device batch.
type DeviceBatch struct {
	ID       string
	Owner    string
	Comment  string
	Group    string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceBatch) Key() string { return "device/devicebatch/" + x.ID }

// Fields encodes x for storage.
func (x DeviceBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceBatch decodes a stored record.
func ParseDeviceBatch(f codec.Fields) (DeviceBatch, error) {
	var x DeviceBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicebatch: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("device: devicebatch %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("device: devicebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceBatch) String() string {
	return fmt.Sprintf("DeviceBatch(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceBatch) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceBatch) Normalize() DeviceBatch {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDeviceBatchs orders by count descending, then by ID.
func SortDeviceBatchs(xs []DeviceBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceBatchs drops the archived entries, preserving order.
func ActiveDeviceBatchs(xs []DeviceBatch) []DeviceBatch {
	out := make([]DeviceBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceBatchs adds up the count of every active entry.
func SumDeviceBatchs(xs []DeviceBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDeviceBatchs keys the entries by ID, last write winning.
func IndexDeviceBatchs(xs []DeviceBatch) map[string]DeviceBatch {
	out := make(map[string]DeviceBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceBatch builds a valid DeviceBatch, for tests and for the seed data the cron
// binary writes.
func SampleDeviceBatch(id string, at time.Time) DeviceBatch {
	return DeviceBatch{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by device.SampleDeviceBatch",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceRequestSchema is the stored shape of a DeviceRequest.
var DeviceRequestSchema = schema.New("device.devicerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceRequest is one device request.
type DeviceRequest struct {
	ID       string
	Category string
	Name     string
	Label    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceRequest) Key() string { return "device/devicerequest/" + x.ID }

// Fields encodes x for storage.
func (x DeviceRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceRequest decodes a stored record.
func ParseDeviceRequest(f codec.Fields) (DeviceRequest, error) {
	var x DeviceRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicerequest: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("device: devicerequest %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("device: devicerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceRequest) String() string {
	return fmt.Sprintf("DeviceRequest(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceRequest) Normalize() DeviceRequest {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDeviceRequests orders by offset descending, then by ID.
func SortDeviceRequests(xs []DeviceRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceRequests drops the archived entries, preserving order.
func ActiveDeviceRequests(xs []DeviceRequest) []DeviceRequest {
	out := make([]DeviceRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceRequests adds up the offset of every active entry.
func SumDeviceRequests(xs []DeviceRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDeviceRequests keys the entries by ID, last write winning.
func IndexDeviceRequests(xs []DeviceRequest) map[string]DeviceRequest {
	out := make(map[string]DeviceRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceRequest builds a valid DeviceRequest, for tests and for the seed data the cron
// binary writes.
func SampleDeviceRequest(id string, at time.Time) DeviceRequest {
	return DeviceRequest{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Label:    "seeded by device.SampleDeviceRequest",
		Offset:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceSummarySchema is the stored shape of a DeviceSummary.
var DeviceSummarySchema = schema.New("device.devicesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceSummary is one device summary.
type DeviceSummary struct {
	ID       string
	Name     string
	Comment  string
	Group    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceSummary) Key() string { return "device/devicesummary/" + x.ID }

// Fields encodes x for storage.
func (x DeviceSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"comment":  x.Comment,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceSummary decodes a stored record.
func ParseDeviceSummary(f codec.Fields) (DeviceSummary, error) {
	var x DeviceSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicesummary: missing id")
	}
	x.Name = f["name"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: devicesummary %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("device: devicesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceSummary) String() string {
	return fmt.Sprintf("DeviceSummary(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceSummary) Normalize() DeviceSummary {
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeviceSummarys orders by capacity descending, then by ID.
func SortDeviceSummarys(xs []DeviceSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceSummarys drops the archived entries, preserving order.
func ActiveDeviceSummarys(xs []DeviceSummary) []DeviceSummary {
	out := make([]DeviceSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceSummarys adds up the capacity of every active entry.
func SumDeviceSummarys(xs []DeviceSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDeviceSummarys keys the entries by ID, last write winning.
func IndexDeviceSummarys(xs []DeviceSummary) map[string]DeviceSummary {
	out := make(map[string]DeviceSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceSummary builds a valid DeviceSummary, for tests and for the seed data the cron
// binary writes.
func SampleDeviceSummary(id string, at time.Time) DeviceSummary {
	return DeviceSummary{
		ID:       id,
		Name:     "name-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by device.SampleDeviceSummary",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DevicePolicySchema is the stored shape of a DevicePolicy.
var DevicePolicySchema = schema.New("device.devicepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DevicePolicy is one device policy.
type DevicePolicy struct {
	ID       string
	Category string
	Channel  string
	Source   string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DevicePolicy) Key() string { return "device/devicepolicy/" + x.ID }

// Fields encodes x for storage.
func (x DevicePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDevicePolicy decodes a stored record.
func ParseDevicePolicy(f codec.Fields) (DevicePolicy, error) {
	var x DevicePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicepolicy: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("device: devicepolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("device: devicepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DevicePolicy) String() string {
	return fmt.Sprintf("DevicePolicy(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DevicePolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DevicePolicy) Normalize() DevicePolicy {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDevicePolicys orders by weight descending, then by ID.
func SortDevicePolicys(xs []DevicePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDevicePolicys drops the archived entries, preserving order.
func ActiveDevicePolicys(xs []DevicePolicy) []DevicePolicy {
	out := make([]DevicePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDevicePolicys adds up the weight of every active entry.
func SumDevicePolicys(xs []DevicePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDevicePolicys keys the entries by ID, last write winning.
func IndexDevicePolicys(xs []DevicePolicy) map[string]DevicePolicy {
	out := make(map[string]DevicePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDevicePolicy builds a valid DevicePolicy, for tests and for the seed data the cron
// binary writes.
func SampleDevicePolicy(id string, at time.Time) DevicePolicy {
	return DevicePolicy{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by device.SampleDevicePolicy",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceLinkSchema is the stored shape of a DeviceLink.
var DeviceLinkSchema = schema.New("device.devicelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceLink is one device link.
type DeviceLink struct {
	ID       string
	Owner    string
	Comment  string
	Channel  string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceLink) Key() string { return "device/devicelink/" + x.ID }

// Fields encodes x for storage.
func (x DeviceLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceLink decodes a stored record.
func ParseDeviceLink(f codec.Fields) (DeviceLink, error) {
	var x DeviceLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicelink: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: devicelink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("device: devicelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceLink) String() string {
	return fmt.Sprintf("DeviceLink(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceLink) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceLink) Normalize() DeviceLink {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeviceLinks orders by capacity descending, then by ID.
func SortDeviceLinks(xs []DeviceLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceLinks drops the archived entries, preserving order.
func ActiveDeviceLinks(xs []DeviceLink) []DeviceLink {
	out := make([]DeviceLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceLinks adds up the capacity of every active entry.
func SumDeviceLinks(xs []DeviceLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDeviceLinks keys the entries by ID, last write winning.
func IndexDeviceLinks(xs []DeviceLink) map[string]DeviceLink {
	out := make(map[string]DeviceLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceLink builds a valid DeviceLink, for tests and for the seed data the cron
// binary writes.
func SampleDeviceLink(id string, at time.Time) DeviceLink {
	return DeviceLink{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Channel:  "seeded by device.SampleDeviceLink",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceNoteSchema is the stored shape of a DeviceNote.
var DeviceNoteSchema = schema.New("device.devicenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceNote is one device note.
type DeviceNote struct {
	ID        string
	Label     string
	Source    string
	Reference string
	Retained  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeviceNote) Key() string { return "device/devicenote/" + x.ID }

// Fields encodes x for storage.
func (x DeviceNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"source":    x.Source,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceNote decodes a stored record.
func ParseDeviceNote(f codec.Fields) (DeviceNote, error) {
	var x DeviceNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicenote: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("device: devicenote %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("device: devicenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceNote) String() string {
	return fmt.Sprintf("DeviceNote(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceNote) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceNote) Normalize() DeviceNote {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDeviceNotes orders by retained descending, then by ID.
func SortDeviceNotes(xs []DeviceNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceNotes drops the archived entries, preserving order.
func ActiveDeviceNotes(xs []DeviceNote) []DeviceNote {
	out := make([]DeviceNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceNotes adds up the retained of every active entry.
func SumDeviceNotes(xs []DeviceNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDeviceNotes keys the entries by ID, last write winning.
func IndexDeviceNotes(xs []DeviceNote) map[string]DeviceNote {
	out := make(map[string]DeviceNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceNote builds a valid DeviceNote, for tests and for the seed data the cron
// binary writes.
func SampleDeviceNote(id string, at time.Time) DeviceNote {
	return DeviceNote{
		ID:        id,
		Label:     "label-" + id,
		Source:    "source-" + id,
		Reference: "seeded by device.SampleDeviceNote",
		Retained:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeviceWindowSchema is the stored shape of a DeviceWindow.
var DeviceWindowSchema = schema.New("device.devicewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceWindow is one device window.
type DeviceWindow struct {
	ID       string
	Group    string
	Category string
	Region   string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceWindow) Key() string { return "device/devicewindow/" + x.ID }

// Fields encodes x for storage.
func (x DeviceWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"region":   x.Region,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceWindow decodes a stored record.
func ParseDeviceWindow(f codec.Fields) (DeviceWindow, error) {
	var x DeviceWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicewindow: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("device: devicewindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("device: devicewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceWindow) String() string {
	return fmt.Sprintf("DeviceWindow(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceWindow) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceWindow) Normalize() DeviceWindow {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDeviceWindows orders by version descending, then by ID.
func SortDeviceWindows(xs []DeviceWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceWindows drops the archived entries, preserving order.
func ActiveDeviceWindows(xs []DeviceWindow) []DeviceWindow {
	out := make([]DeviceWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceWindows adds up the version of every active entry.
func SumDeviceWindows(xs []DeviceWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeviceWindows keys the entries by ID, last write winning.
func IndexDeviceWindows(xs []DeviceWindow) map[string]DeviceWindow {
	out := make(map[string]DeviceWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceWindow builds a valid DeviceWindow, for tests and for the seed data the cron
// binary writes.
func SampleDeviceWindow(id string, at time.Time) DeviceWindow {
	return DeviceWindow{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Region:   "seeded by device.SampleDeviceWindow",
		Version:  len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceQuotaSchema is the stored shape of a DeviceQuota.
var DeviceQuotaSchema = schema.New("device.devicequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceQuota is one device quota.
type DeviceQuota struct {
	ID       string
	Owner    string
	Channel  string
	Category string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceQuota) Key() string { return "device/devicequota/" + x.ID }

// Fields encodes x for storage.
func (x DeviceQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceQuota decodes a stored record.
func ParseDeviceQuota(f codec.Fields) (DeviceQuota, error) {
	var x DeviceQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicequota: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: devicequota %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("device: devicequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceQuota) String() string {
	return fmt.Sprintf("DeviceQuota(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceQuota) Normalize() DeviceQuota {
	x.Owner = strings.TrimSpace(x.Owner)
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

// SortDeviceQuotas orders by capacity descending, then by ID.
func SortDeviceQuotas(xs []DeviceQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceQuotas drops the archived entries, preserving order.
func ActiveDeviceQuotas(xs []DeviceQuota) []DeviceQuota {
	out := make([]DeviceQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceQuotas adds up the capacity of every active entry.
func SumDeviceQuotas(xs []DeviceQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDeviceQuotas keys the entries by ID, last write winning.
func IndexDeviceQuotas(xs []DeviceQuota) map[string]DeviceQuota {
	out := make(map[string]DeviceQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceQuota builds a valid DeviceQuota, for tests and for the seed data the cron
// binary writes.
func SampleDeviceQuota(id string, at time.Time) DeviceQuota {
	return DeviceQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by device.SampleDeviceQuota",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceBucketSchema is the stored shape of a DeviceBucket.
var DeviceBucketSchema = schema.New("device.devicebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceBucket is one device bucket.
type DeviceBucket struct {
	ID       string
	Label    string
	Region   string
	Group    string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceBucket) Key() string { return "device/devicebucket/" + x.ID }

// Fields encodes x for storage.
func (x DeviceBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceBucket decodes a stored record.
func ParseDeviceBucket(f codec.Fields) (DeviceBucket, error) {
	var x DeviceBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicebucket: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("device: devicebucket %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("device: devicebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceBucket) String() string {
	return fmt.Sprintf("DeviceBucket(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceBucket) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceBucket) Normalize() DeviceBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDeviceBuckets orders by weight descending, then by ID.
func SortDeviceBuckets(xs []DeviceBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceBuckets drops the archived entries, preserving order.
func ActiveDeviceBuckets(xs []DeviceBucket) []DeviceBucket {
	out := make([]DeviceBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceBuckets adds up the weight of every active entry.
func SumDeviceBuckets(xs []DeviceBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDeviceBuckets keys the entries by ID, last write winning.
func IndexDeviceBuckets(xs []DeviceBucket) map[string]DeviceBucket {
	out := make(map[string]DeviceBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceBucket builds a valid DeviceBucket, for tests and for the seed data the cron
// binary writes.
func SampleDeviceBucket(id string, at time.Time) DeviceBucket {
	return DeviceBucket{
		ID:      id,
		Label:   "label-" + id,
		Region:  "region-" + id,
		Group:   "seeded by device.SampleDeviceBucket",
		Weight:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// DeviceLaneSchema is the stored shape of a DeviceLane.
var DeviceLaneSchema = schema.New("device.devicelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceLane is one device lane.
type DeviceLane struct {
	ID       string
	Owner    string
	Group    string
	Label    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceLane) Key() string { return "device/devicelane/" + x.ID }

// Fields encodes x for storage.
func (x DeviceLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceLane decodes a stored record.
func ParseDeviceLane(f codec.Fields) (DeviceLane, error) {
	var x DeviceLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicelane: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("device: devicelane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: devicelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceLane) String() string {
	return fmt.Sprintf("DeviceLane(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceLane) Normalize() DeviceLane {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDeviceLanes orders by priority descending, then by ID.
func SortDeviceLanes(xs []DeviceLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceLanes drops the archived entries, preserving order.
func ActiveDeviceLanes(xs []DeviceLane) []DeviceLane {
	out := make([]DeviceLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceLanes adds up the priority of every active entry.
func SumDeviceLanes(xs []DeviceLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeviceLanes keys the entries by ID, last write winning.
func IndexDeviceLanes(xs []DeviceLane) map[string]DeviceLane {
	out := make(map[string]DeviceLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceLane builds a valid DeviceLane, for tests and for the seed data the cron
// binary writes.
func SampleDeviceLane(id string, at time.Time) DeviceLane {
	return DeviceLane{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Label:    "seeded by device.SampleDeviceLane",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceSlotSchema is the stored shape of a DeviceSlot.
var DeviceSlotSchema = schema.New("device.deviceslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceSlot is one device slot.
type DeviceSlot struct {
	ID       string
	Comment  string
	Category string
	Channel  string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceSlot) Key() string { return "device/deviceslot/" + x.ID }

// Fields encodes x for storage.
func (x DeviceSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceSlot decodes a stored record.
func ParseDeviceSlot(f codec.Fields) (DeviceSlot, error) {
	var x DeviceSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: deviceslot: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("device: deviceslot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: deviceslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: deviceslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: deviceslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceSlot) String() string {
	return fmt.Sprintf("DeviceSlot(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceSlot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceSlot) Normalize() DeviceSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDeviceSlots orders by attempts descending, then by ID.
func SortDeviceSlots(xs []DeviceSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceSlots drops the archived entries, preserving order.
func ActiveDeviceSlots(xs []DeviceSlot) []DeviceSlot {
	out := make([]DeviceSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceSlots adds up the attempts of every active entry.
func SumDeviceSlots(xs []DeviceSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDeviceSlots keys the entries by ID, last write winning.
func IndexDeviceSlots(xs []DeviceSlot) map[string]DeviceSlot {
	out := make(map[string]DeviceSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceSlot builds a valid DeviceSlot, for tests and for the seed data the cron
// binary writes.
func SampleDeviceSlot(id string, at time.Time) DeviceSlot {
	return DeviceSlot{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Channel:  "seeded by device.SampleDeviceSlot",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceTagSchema is the stored shape of a DeviceTag.
var DeviceTagSchema = schema.New("device.devicetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceTag is one device tag.
type DeviceTag struct {
	ID       string
	Channel  string
	Category string
	Group    string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceTag) Key() string { return "device/devicetag/" + x.ID }

// Fields encodes x for storage.
func (x DeviceTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceTag decodes a stored record.
func ParseDeviceTag(f codec.Fields) (DeviceTag, error) {
	var x DeviceTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicetag: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("device: devicetag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("device: devicetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceTag) String() string {
	return fmt.Sprintf("DeviceTag(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceTag) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceTag) Normalize() DeviceTag {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDeviceTags orders by attempts descending, then by ID.
func SortDeviceTags(xs []DeviceTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceTags drops the archived entries, preserving order.
func ActiveDeviceTags(xs []DeviceTag) []DeviceTag {
	out := make([]DeviceTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceTags adds up the attempts of every active entry.
func SumDeviceTags(xs []DeviceTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDeviceTags keys the entries by ID, last write winning.
func IndexDeviceTags(xs []DeviceTag) map[string]DeviceTag {
	out := make(map[string]DeviceTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceTag builds a valid DeviceTag, for tests and for the seed data the cron
// binary writes.
func SampleDeviceTag(id string, at time.Time) DeviceTag {
	return DeviceTag{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Group:    "seeded by device.SampleDeviceTag",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeviceRuleSchema is the stored shape of a DeviceRule.
var DeviceRuleSchema = schema.New("device.devicerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceRule is one device rule.
type DeviceRule struct {
	ID        string
	Owner     string
	Reference string
	Group     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeviceRule) Key() string { return "device/devicerule/" + x.ID }

// Fields encodes x for storage.
func (x DeviceRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"reference": x.Reference,
		"group":     x.Group,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceRule decodes a stored record.
func ParseDeviceRule(f codec.Fields) (DeviceRule, error) {
	var x DeviceRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicerule: missing id")
	}
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("device: devicerule %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("device: devicerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceRule) String() string {
	return fmt.Sprintf("DeviceRule(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceRule) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceRule) Normalize() DeviceRule {
	x.Owner = strings.TrimSpace(x.Owner)
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

// SortDeviceRules orders by weight descending, then by ID.
func SortDeviceRules(xs []DeviceRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceRules drops the archived entries, preserving order.
func ActiveDeviceRules(xs []DeviceRule) []DeviceRule {
	out := make([]DeviceRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceRules adds up the weight of every active entry.
func SumDeviceRules(xs []DeviceRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDeviceRules keys the entries by ID, last write winning.
func IndexDeviceRules(xs []DeviceRule) map[string]DeviceRule {
	out := make(map[string]DeviceRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceRule builds a valid DeviceRule, for tests and for the seed data the cron
// binary writes.
func SampleDeviceRule(id string, at time.Time) DeviceRule {
	return DeviceRule{
		ID:        id,
		Owner:     "owner-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by device.SampleDeviceRule",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeviceHoldSchema is the stored shape of a DeviceHold.
var DeviceHoldSchema = schema.New("device.devicehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceHold is one device hold.
type DeviceHold struct {
	ID       string
	Owner    string
	Channel  string
	Group    string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceHold) Key() string { return "device/devicehold/" + x.ID }

// Fields encodes x for storage.
func (x DeviceHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceHold decodes a stored record.
func ParseDeviceHold(f codec.Fields) (DeviceHold, error) {
	var x DeviceHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicehold: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("device: devicehold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("device: devicehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceHold) String() string {
	return fmt.Sprintf("DeviceHold(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceHold) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceHold) Normalize() DeviceHold {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDeviceHolds orders by version descending, then by ID.
func SortDeviceHolds(xs []DeviceHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceHolds drops the archived entries, preserving order.
func ActiveDeviceHolds(xs []DeviceHold) []DeviceHold {
	out := make([]DeviceHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceHolds adds up the version of every active entry.
func SumDeviceHolds(xs []DeviceHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeviceHolds keys the entries by ID, last write winning.
func IndexDeviceHolds(xs []DeviceHold) map[string]DeviceHold {
	out := make(map[string]DeviceHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceHold builds a valid DeviceHold, for tests and for the seed data the cron
// binary writes.
func SampleDeviceHold(id string, at time.Time) DeviceHold {
	return DeviceHold{
		ID:      id,
		Owner:   "owner-" + id,
		Channel: "channel-" + id,
		Group:   "seeded by device.SampleDeviceHold",
		Version: len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// DeviceClaimSchema is the stored shape of a DeviceClaim.
var DeviceClaimSchema = schema.New("device.deviceclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceClaim is one device claim.
type DeviceClaim struct {
	ID        string
	Channel   string
	Reference string
	Name      string
	Capacity  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeviceClaim) Key() string { return "device/deviceclaim/" + x.ID }

// Fields encodes x for storage.
func (x DeviceClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"name":      x.Name,
		"capacity":  strconv.Itoa(x.Capacity),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceClaim decodes a stored record.
func ParseDeviceClaim(f codec.Fields) (DeviceClaim, error) {
	var x DeviceClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: deviceclaim: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("device: deviceclaim %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("device: deviceclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: deviceclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: deviceclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceClaim) String() string {
	return fmt.Sprintf("DeviceClaim(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceClaim) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceClaim) Normalize() DeviceClaim {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDeviceClaims orders by capacity descending, then by ID.
func SortDeviceClaims(xs []DeviceClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceClaims drops the archived entries, preserving order.
func ActiveDeviceClaims(xs []DeviceClaim) []DeviceClaim {
	out := make([]DeviceClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceClaims adds up the capacity of every active entry.
func SumDeviceClaims(xs []DeviceClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDeviceClaims keys the entries by ID, last write winning.
func IndexDeviceClaims(xs []DeviceClaim) map[string]DeviceClaim {
	out := make(map[string]DeviceClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceClaim builds a valid DeviceClaim, for tests and for the seed data the cron
// binary writes.
func SampleDeviceClaim(id string, at time.Time) DeviceClaim {
	return DeviceClaim{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by device.SampleDeviceClaim",
		Capacity:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeviceCursorSchema is the stored shape of a DeviceCursor.
var DeviceCursorSchema = schema.New("device.devicecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeviceCursor is one device cursor.
type DeviceCursor struct {
	ID       string
	Name     string
	Label    string
	Comment  string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeviceCursor) Key() string { return "device/devicecursor/" + x.ID }

// Fields encodes x for storage.
func (x DeviceCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeviceCursor decodes a stored record.
func ParseDeviceCursor(f codec.Fields) (DeviceCursor, error) {
	var x DeviceCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("device: devicecursor: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("device: devicecursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("device: devicecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("device: devicecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("device: devicecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeviceCursor) String() string {
	return fmt.Sprintf("DeviceCursor(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeviceCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeviceCursor) Normalize() DeviceCursor {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeviceCursors orders by version descending, then by ID.
func SortDeviceCursors(xs []DeviceCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeviceCursors drops the archived entries, preserving order.
func ActiveDeviceCursors(xs []DeviceCursor) []DeviceCursor {
	out := make([]DeviceCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeviceCursors adds up the version of every active entry.
func SumDeviceCursors(xs []DeviceCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeviceCursors keys the entries by ID, last write winning.
func IndexDeviceCursors(xs []DeviceCursor) map[string]DeviceCursor {
	out := make(map[string]DeviceCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeviceCursor builds a valid DeviceCursor, for tests and for the seed data the cron
// binary writes.
func SampleDeviceCursor(id string, at time.Time) DeviceCursor {
	return DeviceCursor{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by device.SampleDeviceCursor",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	DeviceRecordSchema,
	DeviceEntrySchema,
	DeviceItemSchema,
	DeviceDraftSchema,
	DeviceRevisionSchema,
	DeviceSnapshotSchema,
	DeviceBatchSchema,
	DeviceRequestSchema,
	DeviceSummarySchema,
	DevicePolicySchema,
	DeviceLinkSchema,
	DeviceNoteSchema,
	DeviceWindowSchema,
	DeviceQuotaSchema,
	DeviceBucketSchema,
	DeviceLaneSchema,
	DeviceSlotSchema,
	DeviceTagSchema,
	DeviceRuleSchema,
	DeviceHoldSchema,
	DeviceClaimSchema,
	DeviceCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"device/devicerecord/",
	"device/deviceentry/",
	"device/deviceitem/",
	"device/devicedraft/",
	"device/devicerevision/",
	"device/devicesnapshot/",
	"device/devicebatch/",
	"device/devicerequest/",
	"device/devicesummary/",
	"device/devicepolicy/",
	"device/devicelink/",
	"device/devicenote/",
	"device/devicewindow/",
	"device/devicequota/",
	"device/devicebucket/",
	"device/devicelane/",
	"device/deviceslot/",
	"device/devicetag/",
	"device/devicerule/",
	"device/devicehold/",
	"device/deviceclaim/",
	"device/devicecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "device: " + strings.Join(names, ", ")
}
