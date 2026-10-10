package payment

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/refund"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/settlement"
)

// Linked returns the summary of every domain package payment builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"payment: " + Summary(),
		"settlement: " + settlement.Summary(),
		"refund: " + refund.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(settlement.Schemas)
	n += len(refund.Schemas)
	return n
}

// LinkedNames lists the domain packages payment links to.
func LinkedNames() []string {
	return strings.Split("settlement,refund", ",")
}
