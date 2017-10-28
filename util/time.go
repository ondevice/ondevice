package util

import "time"

// TsToMsec -- Converts a Time object to its unix timestamp equivalent (with msec granularity)
func TsToMsec(ts time.Time) int64 {
	return ts.UnixNano() / 1000000
}

// MsecToTs -- Convert unix timestamp (with msec granularity) to Time object
func MsecToTs(ts int64) time.Time {
	return time.Unix(ts/1000, (ts*1000000)%1000000000)
}
