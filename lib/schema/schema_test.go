package schema

import (
	"testing"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
)

var testSchema = New("widget",
	Field{Name: "id", Kind: KindString, Required: true, MaxLen: 64},
	Field{Name: "count", Kind: KindInt},
	Field{Name: "active", Kind: KindBool},
)

func TestFieldLookup(t *testing.T) {
	if f, ok := testSchema.Field("count"); !ok || f.Kind != KindInt {
		t.Fatalf("Field(count) = %+v, %v", f, ok)
	}
	if _, ok := testSchema.Field("nope"); ok {
		t.Error("Field(nope) reported present")
	}
}

func TestRequiredAndUnknown(t *testing.T) {
	if got := testSchema.Required(); len(got) != 1 || got[0] != "id" {
		t.Errorf("Required() = %v", got)
	}
	got := testSchema.Unknown(codec.Fields{"id": "x", "zzz": "1", "aaa": "2"})
	if len(got) != 2 || got[0] != "aaa" || got[1] != "zzz" {
		t.Errorf("Unknown() = %v", got)
	}
}

func TestFingerprintIsStable(t *testing.T) {
	want := "widget;id:string!;count:int;active:bool"
	if got := testSchema.Fingerprint(); got != want {
		t.Errorf("Fingerprint() = %q, want %q", got, want)
	}
}

func TestNewPanicsOnDuplicate(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New did not panic on a duplicate field")
		}
	}()
	New("dup", Field{Name: "a"}, Field{Name: "a"})
}
