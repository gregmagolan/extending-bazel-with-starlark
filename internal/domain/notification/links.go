package notification

import (
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/digest"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/template"
)

// Linked returns the summary of every domain package notification builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"notification: " + Summary(),
		"template: " + template.Summary(),
		"digest: " + digest.Summary(),
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
	n += len(template.Schemas)
	n += len(digest.Schemas)
	return n
}

// LinkedNames lists the domain packages notification links to.
func LinkedNames() []string {
	return strings.Split("template,digest", ",")
}
