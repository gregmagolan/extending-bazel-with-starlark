package auth

import (
	"errors"
	"testing"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

func newVerifier(t *testing.T, text, secret string) *Verifier {
	t.Helper()
	cfg, err := config.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	v, err := NewVerifier(log.Discard(), cfg, secret)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMintVerifyRoundTrip(t *testing.T) {
	v := newVerifier(t, "", "s3cret")
	want := Identity{Subject: "u-1", Tenant: "acme", Scopes: []string{"read", "write"}}
	got, err := v.Verify(v.Mint(want))
	if err != nil {
		t.Fatalf("Verify() = %v", err)
	}
	if got.Subject != want.Subject || got.Tenant != want.Tenant || len(got.Scopes) != 2 {
		t.Fatalf("Verify() = %+v, want %+v", got, want)
	}
	if !got.HasScope("write") || got.HasScope("delete") {
		t.Errorf("scopes = %v", got.Scopes)
	}
}

func TestAdminImpliesEveryScope(t *testing.T) {
	id := Identity{Scopes: []string{"admin"}}
	if !id.HasScope("anything") {
		t.Error("admin did not imply scope")
	}
}

func TestTamperedTokenRejected(t *testing.T) {
	v := newVerifier(t, "", "s3cret")
	token := v.Mint(Identity{Subject: "u-1"})
	if _, err := v.Verify(token + "x"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Verify(tampered) = %v", err)
	}
	if _, err := v.Verify("nodot"); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Verify(malformed) = %v", err)
	}
}

func TestOtherSecretRejected(t *testing.T) {
	a := newVerifier(t, "", "one")
	b := newVerifier(t, "", "two")
	if _, err := b.Verify(a.Mint(Identity{Subject: "u-1"})); !errors.Is(err, ErrUnauthorized) {
		t.Errorf("cross-secret verify = %v", err)
	}
}

func TestAnonymousWhenNotRequired(t *testing.T) {
	v := newVerifier(t, "auth.required=false", "")
	id, err := v.Verify("")
	if err != nil || id.Subject != "anonymous" {
		t.Fatalf("Verify(empty) = %+v, %v", id, err)
	}
}

func TestSecretRequiredWhenAuthRequired(t *testing.T) {
	cfg, _ := config.Parse("auth.required=true")
	if _, err := NewVerifier(log.Discard(), cfg, ""); err == nil {
		t.Error("NewVerifier accepted an empty secret")
	}
}
