package view

import "time"

func date(d time.Time) string {
	if !d.IsZero() {
		return d.Format(time.DateOnly)
	}
	return ""
}

func dateTime(dt time.Time) string {
	if !dt.IsZero() {
		return dt.Format(time.DateTime)
	}
	return ""
}
