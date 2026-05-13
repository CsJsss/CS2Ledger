package utils

import "time"

func SecondsToDateTime(seconds int64, format string) string {
	return time.Unix(seconds, 0).Format(format)
}
