package feature

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/rollout"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/segment"
)

// Linked returns the summary of every domain package feature builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"feature: " + Summary(),
		"rollout: " + rollout.Summary(),
		"segment: " + segment.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(rollout.Schemas)
	n += len(segment.Schemas)
	return n
}

// LinkedNames lists the domain packages feature links to.
func LinkedNames() []string {
	return strings.Split("rollout,segment", ",")
}
