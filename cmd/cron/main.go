// Command cron reports the schedule a deployment should run its periodic jobs
// on.
//
// It is intentionally the thinnest of the three binaries: it reads
// configuration, resolves the next few firing times against the clock, and
// prints them. Keeping its dependency set small is what makes the course's
// reverse-dependency queries interesting -- //cmd/cron does not pull in
// lib/store, lib/queue or lib/api.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/clock"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

func main() {
	settings := flag.String("config", "", "inline key=value configuration")
	count := flag.Int("count", 3, "number of firing times to report per entry")
	flag.Parse()

	if err := run(os.Stdout, clock.System{}, *settings, *count); err != nil {
		fmt.Fprintln(os.Stderr, "cron:", err)
		os.Exit(1)
	}
}

// Entry is one scheduled job.
type Entry struct {
	Name   string
	Every  time.Duration
	Jitter time.Duration
}

// schedule is the built-in table. A later course step moves it into
// configuration; for now it is a var so tests can shadow it.
var schedule = []Entry{
	{Name: "compact-store", Every: 6 * time.Hour},
	{Name: "expire-sessions", Every: 15 * time.Minute, Jitter: time.Minute},
	{Name: "retry-webhooks", Every: 5 * time.Minute, Jitter: 30 * time.Second},
	{Name: "roll-reports", Every: 24 * time.Hour},
}

func run(out io.Writer, clk clock.Clock, settings string, count int) error {
	cfg, err := config.Load(settings)
	if err != nil {
		return err
	}
	if err := cfg.Check(); err != nil {
		return err
	}

	level, err := cfg.LogLevel()
	if err != nil {
		return err
	}
	l := log.New(os.Stderr, level).With("service", "cron")

	plan, err := Plan(clk.Now(), schedule, count)
	if err != nil {
		return err
	}
	for _, line := range plan {
		fmt.Fprintln(out, line)
	}
	l.Info("schedule resolved", "entries", fmt.Sprint(len(schedule)), "lines", fmt.Sprint(len(plan)))
	return nil
}

// Plan resolves the next count firing times for each entry, relative to now.
// Entries fire on wall-clock boundaries of their period, with the jitter
// added afterwards so two entries with the same period do not collide.
func Plan(now time.Time, entries []Entry, count int) ([]string, error) {
	if count < 1 {
		return nil, fmt.Errorf("cron: count %d must be positive", count)
	}

	sorted := make([]Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Every != sorted[j].Every {
			return sorted[i].Every < sorted[j].Every
		}
		return sorted[i].Name < sorted[j].Name
	})

	out := make([]string, 0, len(sorted))
	for _, e := range sorted {
		if e.Every <= 0 {
			return nil, fmt.Errorf("cron: %s: period %s must be positive", e.Name, e.Every)
		}
		times := make([]string, 0, count)
		next := now.UTC().Truncate(e.Every)
		for len(times) < count {
			next = next.Add(e.Every)
			times = append(times, next.Add(e.Jitter).Format(time.RFC3339))
		}
		out = append(out, fmt.Sprintf("%s every=%s next=%s", e.Name, e.Every, strings.Join(times, ",")))
	}
	return out, nil
}
