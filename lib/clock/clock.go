// Package clock provides an injectable wall-clock abstraction.
package clock

import "time"

// Clock reports the current time.
type Clock interface {
	Now() time.Time
}

// System is a Clock backed by time.Now.
type System struct{}

// Now returns the current wall-clock time.
func (System) Now() time.Time { return time.Now() }
