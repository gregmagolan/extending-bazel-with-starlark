package inventory

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/stocktake"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/warehouse"
)

// Linked returns the summary of every domain package inventory builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"inventory: " + Summary(),
		"warehouse: " + warehouse.Summary(),
		"stocktake: " + stocktake.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(warehouse.Schemas)
	n += len(stocktake.Schemas)
	return n
}

// LinkedNames lists the domain packages inventory links to.
func LinkedNames() []string {
	return strings.Split("warehouse,stocktake", ",")
}
