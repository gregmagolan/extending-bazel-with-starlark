// Package config loads service configuration from key=value text and the
// process environment.
//
// Precedence is: defaults, then the loaded text, then the environment. The
// environment wins because that is what a container orchestrator controls.
package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Config is a flat, read-mostly settings bag.
type Config struct {
	values map[string]string
}

// Defaults are the settings every service starts from.
func Defaults() map[string]string {
	return map[string]string{
		"log.level":         "info",
		"store.path":        "/var/lib/course/store",
		"store.max_records": "10000",
		"queue.depth":       "256",
		"queue.lease":       "30s",
		"http.port":         "8080",
		"http.timeout":      "15s",
		"auth.required":     "true",
	}
}

// New returns a Config seeded with Defaults.
func New() *Config {
	return &Config{values: Defaults()}
}

// Parse reads "key = value" lines, ignoring blanks and '#' comments.
func Parse(text string) (*Config, error) {
	c := New()
	for n, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			return nil, fmt.Errorf("config: line %d: expected key=value, got %q", n+1, trimmed)
		}
		key = strings.TrimSpace(key)
		if err := validate.NonEmpty("key", key); err != nil {
			return nil, fmt.Errorf("config: line %d: %w", n+1, err)
		}
		c.values[key] = strings.TrimSpace(value)
	}
	return c, nil
}

// FromEnv overlays any COURSE_-prefixed environment variables. The suffix is
// lowercased and '_' becomes '.', so COURSE_HTTP_PORT sets "http.port".
func (c *Config) FromEnv(environ []string) *Config {
	for _, kv := range environ {
		name, value, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(name, "COURSE_") {
			continue
		}
		key := strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(name, "COURSE_"), "_", "."))
		c.values[key] = value
	}
	return c
}

// Load parses text and then overlays the real process environment.
func Load(text string) (*Config, error) {
	c, err := Parse(text)
	if err != nil {
		return nil, err
	}
	return c.FromEnv(os.Environ()), nil
}

// String returns the value for key, or def when unset.
func (c *Config) String(key, def string) string {
	if v, ok := c.values[key]; ok {
		return v
	}
	return def
}

// Int returns the value for key as an int.
func (c *Config) Int(key string, def int) (int, error) {
	v, ok := c.values[key]
	if !ok {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s: %w", key, err)
	}
	return n, nil
}

// Bool returns the value for key as a bool.
func (c *Config) Bool(key string, def bool) (bool, error) {
	v, ok := c.values[key]
	if !ok {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("config: %s: %w", key, err)
	}
	return b, nil
}

// Duration returns the value for key as a time.Duration.
func (c *Config) Duration(key string, def time.Duration) (time.Duration, error) {
	v, ok := c.values[key]
	if !ok {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s: %w", key, err)
	}
	return d, nil
}

// LogLevel resolves "log.level" into a log.Level.
func (c *Config) LogLevel() (log.Level, error) {
	return log.ParseLevel(c.String("log.level", "info"))
}

// Keys returns every configured key, sorted.
func (c *Config) Keys() []string {
	out := make([]string, 0, len(c.values))
	for k := range c.values {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Check reports the first problem with the resolved configuration. Services
// call it during startup so that a typo fails the process rather than the
// first request.
func (c *Config) Check() error {
	if _, err := c.LogLevel(); err != nil {
		return err
	}
	port, err := c.Int("http.port", 8080)
	if err != nil {
		return err
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("config: http.port %d is out of range", port)
	}
	depth, err := c.Int("queue.depth", 256)
	if err != nil {
		return err
	}
	if depth < 1 {
		return fmt.Errorf("config: queue.depth %d must be positive", depth)
	}
	if _, err := c.Duration("queue.lease", 30*time.Second); err != nil {
		return err
	}
	if _, err := c.Duration("http.timeout", 15*time.Second); err != nil {
		return err
	}
	return validate.NonEmpty("store.path", c.String("store.path", ""))
}
