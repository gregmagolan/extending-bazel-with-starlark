package user

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/preference"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/profile"
)

// Linked returns the summary of every domain package user builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"user: " + Summary(),
		"profile: " + profile.Summary(),
		"preference: " + preference.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(profile.Schemas)
	n += len(preference.Schemas)
	return n
}

// LinkedNames lists the domain packages user links to.
func LinkedNames() []string {
	return strings.Split("profile,preference", ",")
}
