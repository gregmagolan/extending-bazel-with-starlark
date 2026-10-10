package clock

import "testing"

func TestSystemNowAdvances(t *testing.T) {
	c := System{}
	if c.Now().IsZero() {
		t.Fatal("System.Now returned the zero time")
	}
}
