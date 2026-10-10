package config

import (
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

func TestParseOverridesDefaults(t *testing.T) {
	c, err := Parse("# a comment\n\nhttp.port = 9090\nlog.level=debug\n")
	if err != nil {
		t.Fatalf("Parse() = %v", err)
	}
	if got, _ := c.Int("http.port", 0); got != 9090 {
		t.Errorf("http.port = %d", got)
	}
	lvl, err := c.LogLevel()
	if err != nil || lvl != log.LevelDebug {
		t.Errorf("LogLevel() = %v, %v", lvl, err)
	}
	if got := c.String("store.path", ""); got == "" {
		t.Error("defaults were dropped")
	}
}

func TestParseRejectsBadLine(t *testing.T) {
	if _, err := Parse("just-a-word\n"); err == nil {
		t.Error("Parse accepted a line with no '='")
	}
	if _, err := Parse("  = value\n"); err == nil {
		t.Error("Parse accepted an empty key")
	}
}

func TestEnvWins(t *testing.T) {
	c, err := Parse("http.port=1111")
	if err != nil {
		t.Fatal(err)
	}
	c.FromEnv([]string{"PATH=/bin", "COURSE_HTTP_PORT=2222", "COURSE_QUEUE_LEASE=5s"})
	if got, _ := c.Int("http.port", 0); got != 2222 {
		t.Errorf("http.port = %d, want 2222", got)
	}
	if got, _ := c.Duration("queue.lease", 0); got != 5*time.Second {
		t.Errorf("queue.lease = %s", got)
	}
}

func TestCheck(t *testing.T) {
	if err := New().Check(); err != nil {
		t.Fatalf("defaults failed Check(): %v", err)
	}
	bad, _ := Parse("http.port=70000")
	if err := bad.Check(); err == nil {
		t.Error("Check accepted an out-of-range port")
	}
	bad, _ = Parse("queue.lease=soon")
	if err := bad.Check(); err == nil {
		t.Error("Check accepted an unparseable duration")
	}
}

func TestKeysAreSorted(t *testing.T) {
	keys := New().Keys()
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("Keys() not sorted at %d: %v", i, keys)
		}
	}
}
