// Package validate checks encoded records against their schema.
//
// Errors accumulate: a caller gets every problem with a record in one pass,
// because the api layer turns the whole list into a single 400 response.
package validate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// Problem is one thing wrong with one field.
type Problem struct {
	Field  string
	Detail string
}

// Error implements the error interface.
func (p Problem) Error() string { return p.Field + ": " + p.Detail }

// Result collects the problems found in a single record.
type Result struct {
	Problems []Problem
}

// OK reports whether the record is valid.
func (r *Result) OK() bool { return len(r.Problems) == 0 }

// Err returns nil when the record is valid, and otherwise an error listing
// every problem in field order.
func (r *Result) Err() error {
	if r.OK() {
		return nil
	}
	parts := make([]string, 0, len(r.Problems))
	for _, p := range r.Problems {
		parts = append(parts, p.Error())
	}
	return fmt.Errorf("invalid record: %s", strings.Join(parts, "; "))
}

func (r *Result) add(field, format string, args ...any) {
	r.Problems = append(r.Problems, Problem{Field: field, Detail: fmt.Sprintf(format, args...)})
}

// Record checks f against s, in the field order declared by the schema.
// Unknown keys are reported, not ignored: a typo in a field name is the most
// common cause of a silently dropped value.
func Record(s schema.Schema, f codec.Fields) *Result {
	res := &Result{}
	for _, field := range s.Fields {
		raw, present := f[field.Name]
		if !present {
			if field.Required {
				res.add(field.Name, "required field is missing")
			}
			continue
		}
		if field.Required && strings.TrimSpace(raw) == "" {
			res.add(field.Name, "required field is blank")
			continue
		}
		if raw == "" {
			continue
		}
		checkKind(res, field, raw)
	}
	for _, k := range s.Unknown(f) {
		res.add(k, "unknown field for schema %q", s.Name)
	}
	return res
}

func checkKind(res *Result, field schema.Field, raw string) {
	switch field.Kind {
	case schema.KindString:
		if field.MaxLen > 0 && len(raw) > field.MaxLen {
			res.add(field.Name, "length %d exceeds maximum %d", len(raw), field.MaxLen)
		}
	case schema.KindInt:
		if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
			res.add(field.Name, "%q is not an integer", raw)
		}
	case schema.KindBool:
		if _, err := strconv.ParseBool(raw); err != nil {
			res.add(field.Name, "%q is not a boolean", raw)
		}
	case schema.KindTimestamp:
		if _, err := time.Parse(time.RFC3339, raw); err != nil {
			res.add(field.Name, "%q is not an RFC3339 timestamp", raw)
		}
	default:
		res.add(field.Name, "unsupported kind %s", field.Kind)
	}
}

// NonEmpty is a convenience check for the hand-written call sites that do not
// have a schema handy.
func NonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return Problem{Field: field, Detail: "must not be empty"}
	}
	return nil
}
