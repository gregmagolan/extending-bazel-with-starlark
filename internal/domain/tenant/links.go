package tenant

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/locale"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/quota"
)

// Linked returns the summary of every domain package tenant builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"tenant: " + Summary(),
		"locale: " + locale.Summary(),
		"quota: " + quota.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(locale.Schemas)
	n += len(quota.Schemas)
	return n
}

// LinkedNames lists the domain packages tenant links to.
func LinkedNames() []string {
	return strings.Split("locale,quota", ",")
}
