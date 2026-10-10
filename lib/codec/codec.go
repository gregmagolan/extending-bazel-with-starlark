// Package codec encodes and decodes the small, self-describing records that
// every other package in this repo passes around.
//
// The encoding is deliberately boring: a sorted, newline-delimited list of
// "key=value" pairs with percent-escaping for the few bytes that would
// otherwise be ambiguous. It exists so that store, queue and api can agree on
// a wire format without pulling in a dependency.
package codec

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Fields is an unordered bag of string-keyed values.
type Fields map[string]string

// Encode renders fields as a deterministic, newline-delimited string.
// Keys are emitted in sorted order so that the output is byte-stable and can
// be used as a cache key.
func Encode(f Fields) string {
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(escape(k))
		b.WriteByte('=')
		b.WriteString(escape(f[k]))
	}
	return b.String()
}

// Decode parses the output of Encode. It rejects input that Encode could not
// have produced, rather than silently repairing it.
func Decode(s string) (Fields, error) {
	out := Fields{}
	if s == "" {
		return out, nil
	}
	for n, line := range strings.Split(s, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("codec: line %d: missing '='", n+1)
		}
		key, err := unescape(k)
		if err != nil {
			return nil, fmt.Errorf("codec: line %d: key: %w", n+1, err)
		}
		val, err := unescape(v)
		if err != nil {
			return nil, fmt.Errorf("codec: line %d: value: %w", n+1, err)
		}
		if _, dup := out[key]; dup {
			return nil, fmt.Errorf("codec: line %d: duplicate key %q", n+1, key)
		}
		out[key] = val
	}
	return out, nil
}

// Int reads a field as an int, returning def when the key is absent.
func (f Fields) Int(key string, def int) (int, error) {
	raw, ok := f[key]
	if !ok {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("codec: field %q: %w", key, err)
	}
	return n, nil
}

// Bool reads a field as a bool, returning def when the key is absent.
func (f Fields) Bool(key string, def bool) (bool, error) {
	raw, ok := f[key]
	if !ok {
		return def, nil
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("codec: field %q: %w", key, err)
	}
	return b, nil
}

// Clone returns a deep copy, so callers can hand a record to a goroutine
// without sharing the map.
func (f Fields) Clone() Fields {
	out := make(Fields, len(f))
	for k, v := range f {
		out[k] = v
	}
	return out
}

func escape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '%':
			b.WriteString("%25")
		case '=':
			b.WriteString("%3D")
		case '\n':
			b.WriteString("%0A")
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func unescape(s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			b.WriteByte(s[i])
			continue
		}
		if i+2 >= len(s) {
			return "", fmt.Errorf("truncated escape at offset %d", i)
		}
		n, err := strconv.ParseUint(s[i+1:i+3], 16, 8)
		if err != nil {
			return "", fmt.Errorf("bad escape %q at offset %d", s[i:i+3], i)
		}
		b.WriteByte(byte(n))
		i += 2
	}
	return b.String(), nil
}
