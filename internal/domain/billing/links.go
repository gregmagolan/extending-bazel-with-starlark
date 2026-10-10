package billing

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/dunning"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/usage"
)

// Linked returns the summary of every domain package billing builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"billing: " + Summary(),
		"dunning: " + dunning.Summary(),
		"usage: " + usage.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(dunning.Schemas)
	n += len(usage.Schemas)
	return n
}

// LinkedNames lists the domain packages billing links to.
func LinkedNames() []string {
	return strings.Split("dunning,usage", ",")
}
