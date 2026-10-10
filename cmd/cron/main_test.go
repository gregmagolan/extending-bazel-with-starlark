package main

import (
	"strings"
	"testing"
	"time"
)

func TestPlanIsSortedByPeriod(t *testing.T) {
	now := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	got, err := Plan(now, schedule, 2)
	if err != nil {
		t.Fatalf("Plan() = %v", err)
	}
	if len(got) != len(schedule) {
		t.Fatalf("Plan() returned %d lines, want %d", len(got), len(schedule))
	}
	if !strings.HasPrefix(got[0], "retry-webhooks") {
		t.Errorf("shortest period is not first: %q", got[0])
	}
	if !strings.HasPrefix(got[len(got)-1], "roll-reports") {
		t.Errorf("longest period is not last: %q", got[len(got)-1])
	}
}

func TestPlanFiresOnBoundariesWithJitter(t *testing.T) {
	now := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	got, err := Plan(now, []Entry{{Name: "e", Every: 15 * time.Minute, Jitter: time.Minute}}, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := "e every=15m0s next=2024-05-06T07:16:00Z,2024-05-06T07:31:00Z"
	if got[0] != want {
		t.Errorf("Plan() = %q, want %q", got[0], want)
	}
}

func TestPlanRejectsBadInput(t *testing.T) {
	now := time.Unix(0, 0).UTC()
	if _, err := Plan(now, schedule, 0); err == nil {
		t.Error("Plan() accepted a zero count")
	}
	if _, err := Plan(now, []Entry{{Name: "bad"}}, 1); err == nil {
		t.Error("Plan() accepted a zero period")
	}
}

func TestRunRejectsBadConfig(t *testing.T) {
	if err := run(nil, nil, "store.path=", 1); err == nil {
		t.Error("run() accepted an empty store path")
	}
}
