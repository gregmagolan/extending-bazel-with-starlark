package shipping

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/carrier"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/route"
)

// Linked returns the summary of every domain package shipping builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"shipping: " + Summary(),
		"carrier: " + carrier.Summary(),
		"route: " + route.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(carrier.Schemas)
	n += len(route.Schemas)
	return n
}

// LinkedNames lists the domain packages shipping links to.
func LinkedNames() []string {
	return strings.Split("carrier,route", ",")
}
