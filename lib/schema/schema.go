// Package schema describes the shape of a record: which fields exist, which
// are required, and what primitive kind each one holds.
//
// A Schema is data, not code, so the domain packages can declare their shape
// once and have validate, store and api all agree on it.
package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
)

// Kind enumerates the primitive field kinds the encoding supports.
type Kind int

// The supported field kinds.
const (
	KindString Kind = iota
	KindInt
	KindBool
	KindTimestamp
)

// String implements fmt.Stringer.
func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindInt:
		return "int"
	case KindBool:
		return "bool"
	case KindTimestamp:
		return "timestamp"
	default:
		return fmt.Sprintf("kind(%d)", int(k))
	}
}

// Field is one named slot in a Schema.
type Field struct {
	Name     string
	Kind     Kind
	Required bool
	// MaxLen, when non-zero, bounds the encoded length of a string field.
	MaxLen int
}

// Schema is an ordered set of fields identified by a stable name.
type Schema struct {
	Name   string
	Fields []Field
}

// New builds a Schema and panics if two fields share a name. Schemas are
// declared in package-level vars, so a panic here fails at init time rather
// than on the first request.
func New(name string, fields ...Field) Schema {
	seen := make(map[string]bool, len(fields))
	for _, f := range fields {
		if seen[f.Name] {
			panic(fmt.Sprintf("schema %q: duplicate field %q", name, f.Name))
		}
		seen[f.Name] = true
	}
	return Schema{Name: name, Fields: fields}
}

// Field looks up a field by name.
func (s Schema) Field(name string) (Field, bool) {
	for _, f := range s.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// Required returns the names of the required fields, sorted.
func (s Schema) Required() []string {
	var out []string
	for _, f := range s.Fields {
		if f.Required {
			out = append(out, f.Name)
		}
	}
	sort.Strings(out)
	return out
}

// Unknown returns the keys of f that the schema does not declare, sorted.
func (s Schema) Unknown(f codec.Fields) []string {
	var out []string
	for k := range f {
		if _, ok := s.Field(k); !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// Fingerprint is a stable, human-readable identity for the schema. Changing a
// field changes the fingerprint, which is how store detects stale records.
func (s Schema) Fingerprint() string {
	var b strings.Builder
	b.WriteString(s.Name)
	for _, f := range s.Fields {
		fmt.Fprintf(&b, ";%s:%s", f.Name, f.Kind)
		if f.Required {
			b.WriteString("!")
		}
	}
	return b.String()
}
