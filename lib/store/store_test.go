package store

import (
	"errors"
	"testing"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

func open(t *testing.T, text string) *Store {
	t.Helper()
	cfg, err := config.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(log.Discard(), cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPutGetRevisions(t *testing.T) {
	s := open(t, "")
	rev, err := s.Put("a", 0, codec.Fields{"x": "1"})
	if err != nil || rev != 1 {
		t.Fatalf("Put() = %d, %v", rev, err)
	}
	if rev, err = s.Put("a", 1, codec.Fields{"x": "2"}); err != nil || rev != 2 {
		t.Fatalf("Put(expect=1) = %d, %v", rev, err)
	}
	if _, err := s.Put("a", 1, codec.Fields{"x": "3"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale Put() = %v, want ErrConflict", err)
	}
	got, err := s.Get("a")
	if err != nil || got.Fields["x"] != "2" {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
}

func TestGetReturnsACopy(t *testing.T) {
	s := open(t, "")
	if _, err := s.Put("a", 0, codec.Fields{"x": "1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("a")
	got.Fields["x"] = "mutated"
	again, _ := s.Get("a")
	if again.Fields["x"] != "1" {
		t.Error("mutating a returned record changed the store")
	}
}

func TestNotFoundAndDelete(t *testing.T) {
	s := open(t, "")
	if _, err := s.Get("missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(missing) = %v", err)
	}
	if err := s.Delete("missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete(missing) = %v", err)
	}
	if _, err := s.Put("a", 0, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("a"); err != nil {
		t.Fatalf("Delete() = %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("Len() = %d, want 0", s.Len())
	}
}

func TestCapacity(t *testing.T) {
	s := open(t, "store.max_records=2")
	for _, k := range []string{"a", "b"} {
		if _, err := s.Put(k, 0, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Put("c", 0, nil); !errors.Is(err, ErrFull) {
		t.Fatalf("Put() over capacity = %v, want ErrFull", err)
	}
	if _, err := s.Put("a", 1, nil); err != nil {
		t.Errorf("overwrite at capacity = %v", err)
	}
}

func TestSchemaIsEnforcedByPrefix(t *testing.T) {
	s := open(t, "")
	s.Register("user/", schema.New("user",
		schema.Field{Name: "email", Kind: schema.KindString, Required: true},
	))
	if _, err := s.Put("user/1", 0, codec.Fields{"email": "a@b.c"}); err != nil {
		t.Fatalf("valid Put() = %v", err)
	}
	if _, err := s.Put("user/2", 0, codec.Fields{}); err == nil {
		t.Error("Put accepted a record missing a required field")
	}
	if _, err := s.Put("other/1", 0, codec.Fields{}); err != nil {
		t.Errorf("unregistered prefix was validated: %v", err)
	}
}

func TestListAndKeysAreSorted(t *testing.T) {
	s := open(t, "")
	for _, k := range []string{"b/2", "a/1", "b/1"} {
		if _, err := s.Put(k, 0, nil); err != nil {
			t.Fatal(err)
		}
	}
	got := s.List("b/")
	if len(got) != 2 || got[0].Key != "b/1" || got[1].Key != "b/2" {
		t.Fatalf("List() = %+v", got)
	}
	if keys := s.Keys(); len(keys) != 3 || keys[0] != "a/1" {
		t.Fatalf("Keys() = %v", keys)
	}
	if len(s.Stats()) == 0 {
		t.Error("Stats() was empty")
	}
}

func TestOpenRejectsBadCapacity(t *testing.T) {
	cfg, _ := config.Parse("store.max_records=0")
	if _, err := Open(log.Discard(), cfg, nil); err == nil {
		t.Error("Open accepted a zero capacity")
	}
}
