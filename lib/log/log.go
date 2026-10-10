// Package log is the structured logger every service in this repo writes to.
//
// It is intentionally small: a level, a set of sticky fields, and a writer.
// The clock is injected so that tests can assert on the rendered line.
package log

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/clock"
)

// Level is a severity threshold.
type Level int

// The supported severities, in increasing order.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String implements fmt.Stringer.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return fmt.Sprintf("level(%d)", int(l))
	}
}

// ParseLevel maps a configuration string onto a Level.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug, nil
	case "info", "":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("log: unknown level %q", s)
	}
}

// Logger writes levelled, structured lines to an io.Writer.
//
// A Logger is safe for concurrent use. With returns a child that shares the
// parent's writer and mutex, so interleaved writes stay line-atomic.
type Logger struct {
	mu     *sync.Mutex
	w      io.Writer
	clk    clock.Clock
	level  Level
	fields map[string]string
}

// New returns a Logger writing to w at the given level.
func New(w io.Writer, level Level) *Logger {
	return &Logger{mu: &sync.Mutex{}, w: w, clk: clock.System{}, level: level}
}

// Default returns a Logger writing to stderr at info level.
func Default() *Logger { return New(os.Stderr, LevelInfo) }

// Discard returns a Logger that drops everything, for tests and benchmarks.
func Discard() *Logger { return New(io.Discard, LevelError) }

// WithClock returns a copy that reads time from clk.
func (l *Logger) WithClock(clk clock.Clock) *Logger {
	out := l.clone()
	out.clk = clk
	return out
}

// With returns a child logger that stamps every line with the given fields.
// Keys already present on the parent are overwritten.
func (l *Logger) With(kv ...string) *Logger {
	if len(kv)%2 != 0 {
		kv = append(kv, "")
	}
	out := l.clone()
	for i := 0; i < len(kv); i += 2 {
		out.fields[kv[i]] = kv[i+1]
	}
	return out
}

// Level reports the logger's current threshold.
func (l *Logger) Level() Level { return l.level }

// Enabled reports whether a line at lvl would be written.
func (l *Logger) Enabled(lvl Level) bool { return lvl >= l.level }

// Debug, Info, Warn and Error write one line at the named severity.
func (l *Logger) Debug(msg string, kv ...string) { l.log(LevelDebug, msg, kv) }
func (l *Logger) Info(msg string, kv ...string)  { l.log(LevelInfo, msg, kv) }
func (l *Logger) Warn(msg string, kv ...string)  { l.log(LevelWarn, msg, kv) }
func (l *Logger) Error(msg string, kv ...string) { l.log(LevelError, msg, kv) }

func (l *Logger) clone() *Logger {
	fields := make(map[string]string, len(l.fields)+2)
	for k, v := range l.fields {
		fields[k] = v
	}
	return &Logger{mu: l.mu, w: l.w, clk: l.clk, level: l.level, fields: fields}
}

func (l *Logger) log(lvl Level, msg string, kv []string) {
	if !l.Enabled(lvl) {
		return
	}
	if len(kv)%2 != 0 {
		kv = append(kv, "")
	}

	merged := make(map[string]string, len(l.fields)+len(kv)/2)
	for k, v := range l.fields {
		merged[k] = v
	}
	for i := 0; i < len(kv); i += 2 {
		merged[kv[i]] = kv[i+1]
	}
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(l.clk.Now().UTC().Format(time.RFC3339))
	b.WriteByte(' ')
	b.WriteString(lvl.String())
	b.WriteByte(' ')
	b.WriteString(msg)
	for _, k := range keys {
		fmt.Fprintf(&b, " %s=%s", k, merged[k])
	}
	b.WriteByte('\n')

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.w, b.String())
}
