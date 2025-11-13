package utils

import (
	"math"
	"time"
)

const (
	// TimeDriftTolerance is the maximum time difference (in seconds) to consider two timestamps as equal.
	// This accounts for filesystem timestamp precision and minor clock drift.
	TimeDriftTolerance = 3
)

// TimeWithinDrift checks if two timestamps are within the drift tolerance.
// Returns true if the absolute difference is <= TimeDriftTolerance seconds.
func TimeWithinDrift(t1, t2 time.Time) bool {
	diff := math.Abs(float64(t1.Unix() - t2.Unix()))
	return diff <= TimeDriftTolerance
}

// TimeDiffSeconds returns the difference in seconds between t1 and t2 (t1 - t2).
// Positive value means t1 is newer, negative means t2 is newer.
func TimeDiffSeconds(t1, t2 time.Time) int64 {
	return t1.Unix() - t2.Unix()
}

// IsNewerThan checks if t1 is definitively newer than t2, accounting for drift tolerance.
// Returns true only if t1 is more than TimeDriftTolerance seconds newer than t2.
func IsNewerThan(t1, t2 time.Time) bool {
	diff := TimeDiffSeconds(t1, t2)
	return diff > TimeDriftTolerance
}
