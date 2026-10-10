package invoice

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/ledger"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/order"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/pricing"
)

// Linked returns the summary of every domain package invoice builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"invoice: " + Summary(),
		"order: " + order.Summary(),
		"pricing: " + pricing.Summary(),
		"ledger: " + ledger.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(order.Schemas)
	n += len(pricing.Schemas)
	n += len(ledger.Schemas)
	return n
}

// LinkedNames lists the domain packages invoice links to.
func LinkedNames() []string {
	return strings.Split("order,pricing,ledger", ",")
}
