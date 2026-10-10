package validate

import (
	"strings"
	"testing"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

var s = schema.New("widget",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 4},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "active", Kind: schema.KindBool},
	schema.Field{Name: "seen", Kind: schema.KindTimestamp},
)

func TestValidRecord(t *testing.T) {
	res := Record(s, codec.Fields{"id": "abc", "count": "3", "active": "true", "seen": "2024-01-02T03:04:05Z"})
	if !res.OK() {
		t.Fatalf("Record() problems = %v", res.Problems)
	}
	if res.Err() != nil {
		t.Errorf("Err() = %v, want nil", res.Err())
	}
}

func TestProblemsAccumulate(t *testing.T) {
	res := Record(s, codec.Fields{"count": "x", "active": "maybe", "stray": "1"})
	if len(res.Problems) != 4 {
		t.Fatalf("got %d problems, want 4: %v", len(res.Problems), res.Problems)
	}
	if !strings.Contains(res.Err().Error(), "unknown field") {
		t.Errorf("Err() = %v", res.Err())
	}
}

func TestMaxLen(t *testing.T) {
	res := Record(s, codec.Fields{"id": "toolong"})
	if res.OK() {
		t.Fatal("over-long id accepted")
	}
}

func TestNonEmpty(t *testing.T) {
	if err := NonEmpty("name", " "); err == nil {
		t.Error("NonEmpty accepted whitespace")
	}
	if err := NonEmpty("name", "ok"); err != nil {
		t.Errorf("NonEmpty() = %v", err)
	}
}
