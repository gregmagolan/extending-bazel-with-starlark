package codec

import "testing"

func TestEncodeIsSorted(t *testing.T) {
	got := Encode(Fields{"b": "2", "a": "1", "c": "3"})
	want := "a=1\nb=2\nc=3"
	if got != want {
		t.Fatalf("Encode() = %q, want %q", got, want)
	}
}

func TestRoundTripEscapes(t *testing.T) {
	in := Fields{"key=with=eq": "line1\nline2", "pct": "100%"}
	out, err := Decode(Encode(in))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	for k, v := range in {
		if out[k] != v {
			t.Errorf("field %q = %q, want %q", k, out[k], v)
		}
	}
}

func TestDecodeRejectsGarbage(t *testing.T) {
	for _, in := range []string{"nokey", "a=1\na=2", "a=%2"} {
		if _, err := Decode(in); err == nil {
			t.Errorf("Decode(%q) succeeded, want error", in)
		}
	}
}

func TestFieldsIntAndBool(t *testing.T) {
	f := Fields{"n": "7", "on": "true"}
	if n, err := f.Int("n", 0); err != nil || n != 7 {
		t.Errorf("Int(n) = %d, %v", n, err)
	}
	if n, err := f.Int("missing", 42); err != nil || n != 42 {
		t.Errorf("Int(missing) = %d, %v", n, err)
	}
	if b, err := f.Bool("on", false); err != nil || !b {
		t.Errorf("Bool(on) = %v, %v", b, err)
	}
	if _, err := f.Int("on", 0); err == nil {
		t.Error("Int(on) succeeded, want error")
	}
}
