package time

import (
	"time"
)

func ParserTimestampMs(timestamp int64) time.Time {
	return time.Unix(0, timestamp*1000000)
}
