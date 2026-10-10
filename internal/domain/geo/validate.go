package geo

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded geo records against this package's schemas.
//
// The service layer calls Check before a write so that a bad request fails
// with every problem listed at once, rather than one problem per round trip.
type Validator struct{ strict bool }

// NewValidator returns a Validator. In strict mode an unknown schema name is
// an error instead of a pass.
func NewValidator(strict bool) Validator { return Validator{strict: strict} }

// Check validates f against the named schema.
func (v Validator) Check(name string, f codec.Fields) error {
	for _, s := range Schemas {
		if s.Name == name {
			return validate.Record(s, f).Err()
		}
	}
	if v.strict {
		return fmt.Errorf("geo: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against GeoRecordSchema.
func (x GeoRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: georecord: %w", err)
	}
	if err := validate.Record(GeoRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: georecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoRecords returns the first problem found in xs.
func ValidateGeoRecords(xs []GeoRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoEntrySchema.
func (x GeoEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geoentry: %w", err)
	}
	if err := validate.Record(GeoEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geoentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoEntrys returns the first problem found in xs.
func ValidateGeoEntrys(xs []GeoEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoItemSchema.
func (x GeoItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geoitem: %w", err)
	}
	if err := validate.Record(GeoItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geoitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoItems returns the first problem found in xs.
func ValidateGeoItems(xs []GeoItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoDraftSchema.
func (x GeoDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geodraft: %w", err)
	}
	if err := validate.Record(GeoDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geodraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoDrafts returns the first problem found in xs.
func ValidateGeoDrafts(xs []GeoDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoRevisionSchema.
func (x GeoRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: georevision: %w", err)
	}
	if err := validate.Record(GeoRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: georevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoRevisions returns the first problem found in xs.
func ValidateGeoRevisions(xs []GeoRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoSnapshotSchema.
func (x GeoSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geosnapshot: %w", err)
	}
	if err := validate.Record(GeoSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geosnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoSnapshots returns the first problem found in xs.
func ValidateGeoSnapshots(xs []GeoSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoBatchSchema.
func (x GeoBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geobatch: %w", err)
	}
	if err := validate.Record(GeoBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geobatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoBatchs returns the first problem found in xs.
func ValidateGeoBatchs(xs []GeoBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoRequestSchema.
func (x GeoRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: georequest: %w", err)
	}
	if err := validate.Record(GeoRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: georequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoRequests returns the first problem found in xs.
func ValidateGeoRequests(xs []GeoRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoSummarySchema.
func (x GeoSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geosummary: %w", err)
	}
	if err := validate.Record(GeoSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geosummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoSummarys returns the first problem found in xs.
func ValidateGeoSummarys(xs []GeoSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoPolicySchema.
func (x GeoPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geopolicy: %w", err)
	}
	if err := validate.Record(GeoPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geopolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoPolicys returns the first problem found in xs.
func ValidateGeoPolicys(xs []GeoPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoLinkSchema.
func (x GeoLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geolink: %w", err)
	}
	if err := validate.Record(GeoLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geolink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoLinks returns the first problem found in xs.
func ValidateGeoLinks(xs []GeoLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoNoteSchema.
func (x GeoNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geonote: %w", err)
	}
	if err := validate.Record(GeoNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geonote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoNotes returns the first problem found in xs.
func ValidateGeoNotes(xs []GeoNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoWindowSchema.
func (x GeoWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geowindow: %w", err)
	}
	if err := validate.Record(GeoWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geowindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoWindows returns the first problem found in xs.
func ValidateGeoWindows(xs []GeoWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoQuotaSchema.
func (x GeoQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geoquota: %w", err)
	}
	if err := validate.Record(GeoQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geoquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoQuotas returns the first problem found in xs.
func ValidateGeoQuotas(xs []GeoQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoBucketSchema.
func (x GeoBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geobucket: %w", err)
	}
	if err := validate.Record(GeoBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geobucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoBuckets returns the first problem found in xs.
func ValidateGeoBuckets(xs []GeoBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoLaneSchema.
func (x GeoLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geolane: %w", err)
	}
	if err := validate.Record(GeoLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geolane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoLanes returns the first problem found in xs.
func ValidateGeoLanes(xs []GeoLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoSlotSchema.
func (x GeoSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geoslot: %w", err)
	}
	if err := validate.Record(GeoSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geoslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoSlots returns the first problem found in xs.
func ValidateGeoSlots(xs []GeoSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoTagSchema.
func (x GeoTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geotag: %w", err)
	}
	if err := validate.Record(GeoTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geotag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoTags returns the first problem found in xs.
func ValidateGeoTags(xs []GeoTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoRuleSchema.
func (x GeoRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: georule: %w", err)
	}
	if err := validate.Record(GeoRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: georule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoRules returns the first problem found in xs.
func ValidateGeoRules(xs []GeoRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoHoldSchema.
func (x GeoHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geohold: %w", err)
	}
	if err := validate.Record(GeoHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geohold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoHolds returns the first problem found in xs.
func ValidateGeoHolds(xs []GeoHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoClaimSchema.
func (x GeoClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geoclaim: %w", err)
	}
	if err := validate.Record(GeoClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geoclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoClaims returns the first problem found in xs.
func ValidateGeoClaims(xs []GeoClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against GeoCursorSchema.
func (x GeoCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("geo: geocursor: %w", err)
	}
	if err := validate.Record(GeoCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("geo: geocursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateGeoCursors returns the first problem found in xs.
func ValidateGeoCursors(xs []GeoCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
