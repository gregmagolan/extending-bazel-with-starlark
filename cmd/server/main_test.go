package main

import "testing"

func TestRunServesScriptedRequests(t *testing.T) {
	if err := run("log.level=error\nhttp.port=9100", "secret", 2); err != nil {
		t.Fatalf("run() = %v", err)
	}
}

func TestRunRejectsBadConfig(t *testing.T) {
	if err := run("http.port=0", "secret", 1); err == nil {
		t.Error("run() accepted an out-of-range port")
	}
	if err := run("not-a-setting", "secret", 1); err == nil {
		t.Error("run() accepted unparseable configuration")
	}
}

func TestRunRequiresSecretWhenAuthRequired(t *testing.T) {
	if err := run("log.level=error", "", 1); err == nil {
		t.Error("run() accepted an empty secret with auth.required set")
	}
}
