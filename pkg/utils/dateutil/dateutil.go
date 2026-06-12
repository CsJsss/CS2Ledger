// Package dateutil provides shared date formatting and weekday mapping for the application.
package dateutil

import "time"

// DateFormat is the standard date string format used across services.
const DateFormat = "2006-01-02"

// DayOfWeekNames maps time.Weekday (0=Sunday) to Chinese weekday names.
var DayOfWeekNames = [...]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}

// FormatTimestamp converts a Unix-millisecond timestamp to a date string and Chinese weekday name.
func FormatTimestamp(tsMillis int64) (date string, dayOfWeek string) {
	t := time.UnixMilli(tsMillis)
	return t.Format(DateFormat), DayOfWeekNames[t.Weekday()]
}

// ParseDate parses a DateFormat-formatted string into a time.Time.
// It always succeeds for strings produced by FormatTimestamp.
func ParseDate(date string) (time.Time, error) {
	return time.Parse(DateFormat, date)
}
