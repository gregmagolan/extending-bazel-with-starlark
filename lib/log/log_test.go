package log

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func newTestLogger(level Level) (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := New(&buf, level).WithClock(fixedClock{time.Unix(0, 0).UTC()})
	return l, &buf
}

func TestLineFormat(t *testing.T) {
	l, buf := newTestLogger(LevelDebug)
	l.Info("started", "port", "8080")
	want := "1970-01-01T00:00:00Z info started port=8080\n"
	if got := buf.String(); got != want {
		t.Fatalf("line = %q, want %q", got, want)
	}
}

func TestLevelFiltering(t *testing.T) {
	l, buf := newTestLogger(LevelWarn)
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")
	lines := strings.Count(buf.String(), "\n")
	if lines != 2 {
		t.Fatalf("wrote %d lines, want 2: %q", lines, buf.String())
	}
}

func TestWithFieldsAreInherited(t *testing.T) {
	l, buf := newTestLogger(LevelInfo)
	child := l.With("service", "server").With("rev", "abc")
	child.Info("hello", "service", "override")
	got := buf.String()
	if !strings.Contains(got, "rev=abc") || !strings.Contains(got, "service=override") {
		t.Fatalf("line = %q", got)
	}
	buf.Reset()
	l.Info("parent")
	if strings.Contains(buf.String(), "rev=") {
		t.Errorf("child fields leaked to parent: %q", buf.String())
	}
}

func TestParseLevel(t *testing.T) {
	for in, want := range map[string]Level{"debug": LevelDebug, "": LevelInfo, "WARNING": LevelWarn, "error": LevelError} {
		got, err := ParseLevel(in)
		if err != nil || got != want {
			t.Errorf("ParseLevel(%q) = %v, %v", in, got, err)
		}
	}
	if _, err := ParseLevel("loud"); err == nil {
		t.Error("ParseLevel(loud) succeeded")
	}
}

func TestOddFieldCountDoesNotPanic(t *testing.T) {
	l, _ := newTestLogger(LevelInfo)
	l.Info("msg", "lonely")
}
