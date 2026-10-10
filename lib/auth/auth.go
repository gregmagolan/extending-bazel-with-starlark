// Package auth verifies the opaque bearer tokens the api layer accepts.
//
// Tokens are symmetric and self-contained: an encoded record plus a keyed
// checksum. This is not a real signing scheme and is not meant to be one; it
// is a stand-in with the same shape so the dependency graph is realistic.
package auth

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

// ErrUnauthorized is returned for any token that does not check out. Callers
// must not distinguish the reasons to the client.
var ErrUnauthorized = errors.New("auth: unauthorized")

// Identity is who a verified token belongs to.
type Identity struct {
	Subject string
	Tenant  string
	Scopes  []string
}

// HasScope reports whether the identity carries the named scope.
func (i Identity) HasScope(scope string) bool {
	for _, s := range i.Scopes {
		if s == scope || s == "admin" {
			return true
		}
	}
	return false
}

// Verifier mints and checks tokens.
type Verifier struct {
	secret   string
	required bool
	log      *log.Logger
}

// NewVerifier reads "auth.required" from cfg. The secret is a process-level
// input, not configuration, so it is passed separately.
func NewVerifier(l *log.Logger, cfg *config.Config, secret string) (*Verifier, error) {
	required, err := cfg.Bool("auth.required", true)
	if err != nil {
		return nil, err
	}
	if required && secret == "" {
		return nil, errors.New("auth: auth.required is set but no secret was provided")
	}
	return &Verifier{secret: secret, required: required, log: l}, nil
}

// Mint encodes an identity as a token.
func (v *Verifier) Mint(id Identity) string {
	body := codec.Encode(codec.Fields{
		"sub":    id.Subject,
		"tenant": id.Tenant,
		"scopes": strings.Join(id.Scopes, ","),
	})
	return checksum(v.secret, body) + "." + strings.ReplaceAll(body, "\n", ";")
}

// Verify parses and checks a token. When auth.required is false an empty
// token yields the anonymous identity, which is how local development works.
func (v *Verifier) Verify(token string) (Identity, error) {
	if token == "" && !v.required {
		return Identity{Subject: "anonymous", Tenant: "local", Scopes: []string{"admin"}}, nil
	}
	sum, body, ok := strings.Cut(token, ".")
	if !ok {
		return Identity{}, fmt.Errorf("%w: malformed token", ErrUnauthorized)
	}
	body = strings.ReplaceAll(body, ";", "\n")
	if checksum(v.secret, body) != sum {
		v.log.Warn("token rejected", "reason", "checksum")
		return Identity{}, fmt.Errorf("%w: bad checksum", ErrUnauthorized)
	}
	fields, err := codec.Decode(body)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrUnauthorized, err)
	}
	id := Identity{Subject: fields["sub"], Tenant: fields["tenant"]}
	if raw := fields["scopes"]; raw != "" {
		id.Scopes = strings.Split(raw, ",")
	}
	if id.Subject == "" {
		return Identity{}, fmt.Errorf("%w: token has no subject", ErrUnauthorized)
	}
	return id, nil
}

func checksum(secret, body string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(secret))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(body))
	return strconv.FormatUint(h.Sum64(), 36)
}
