package ledger

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/pricing"
)

// Linked returns the summary of every domain package ledger builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"ledger: " + Summary(),
		"pricing: " + pricing.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(pricing.Schemas)
	return n
}

// LinkedNames lists the domain packages ledger links to.
func LinkedNames() []string {
	return strings.Split("pricing", ",")
}
