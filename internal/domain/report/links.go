package report

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/forecast"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/geo"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/ledger"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/locale"
)

// Linked returns the summary of every domain package report builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"report: " + Summary(),
		"ledger: " + ledger.Summary(),
		"geo: " + geo.Summary(),
		"locale: " + locale.Summary(),
		"forecast: " + forecast.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(ledger.Schemas)
	n += len(geo.Schemas)
	n += len(locale.Schemas)
	n += len(forecast.Schemas)
	return n
}

// LinkedNames lists the domain packages report links to.
func LinkedNames() []string {
	return strings.Split("ledger,geo,locale,forecast", ",")
}
