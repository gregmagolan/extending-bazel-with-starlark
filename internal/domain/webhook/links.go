package webhook

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/delivery"
)

// Linked returns the summary of every domain package webhook builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"webhook: " + Summary(),
		"delivery: " + delivery.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(delivery.Schemas)
	return n
}

// LinkedNames lists the domain packages webhook links to.
func LinkedNames() []string {
	return strings.Split("delivery", ",")
}
