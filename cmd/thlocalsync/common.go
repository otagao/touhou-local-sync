package main

import (
	"time"
)

// getCurrentTime returns the current time in UTC.
func getCurrentTime() time.Time {
	return time.Now().UTC()
}
