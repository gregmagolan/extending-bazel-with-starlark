package api

import "time"

// zeroTime is the timestamp the seed records carry. Seeded data is not real
// data, so it gets a fixed, obviously-synthetic time rather than time.Now.
func zeroTime() time.Time {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
}
