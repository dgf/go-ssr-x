package web

import (
	"net/url"
	"strconv"
	"time"
)

func parseDate(date string) time.Time {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return time.Time{}
	}
	return t
}

func param2IntOrDefault(query url.Values, key string, defaultValue int) int {
	if value, err := strconv.Atoi(query.Get(key)); err != nil {
		return defaultValue
	} else {
		return value
	}
}
