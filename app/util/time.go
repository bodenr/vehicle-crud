// Package util provides utility implementations.
package util

import "time"

// TimeMillis returns the current unix time in milliseconds.
func TimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// TimeFromMillis converts from unix time in milliseconds to a Time struct.
func TimeFromMillis(ts int64) time.Time {
	return time.Unix(0, ts*int64(time.Millisecond))
}
