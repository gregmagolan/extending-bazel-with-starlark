package session

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/device"
)

// Linked returns the summary of every domain package session builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"session: " + Summary(),
		"device: " + device.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(device.Schemas)
	return n
}

// LinkedNames lists the domain packages session links to.
func LinkedNames() []string {
	return strings.Split("device", ",")
}
