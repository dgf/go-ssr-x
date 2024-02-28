package server

import "time"

func parseDate(date string) time.Time {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return time.Time{}
	}
	return t
}
