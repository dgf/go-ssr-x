package locale

import (
	"time"

	"golang.org/x/text/language"
)

type formatter struct {
	formatDate     func(d time.Time) string
	formatDateTime func(dt time.Time) string
}

var germanFormatter = &formatter{
	formatDate: func(d time.Time) string {
		return d.Format("02.01.2006")
	},
	formatDateTime: func(dt time.Time) string {
		return dt.Format("02.01.2006 15:04:05")
	},
}

var englishFormatter = &formatter{
	formatDate: func(d time.Time) string {
		return d.Format("01/02/2006")
	},
	formatDateTime: func(dt time.Time) string {
		return dt.Format("01/02/2006 3:04PM")
	},
}

var defaultFormatter = &formatter{
	formatDate: func(d time.Time) string {
		return d.Format(time.DateOnly)
	},
	formatDateTime: func(dt time.Time) string {
		return dt.Format(time.DateTime)
	},
}

func refFormatter(lang language.Tag) *formatter {
	switch lang {
	case language.German:
		return germanFormatter
	case language.AmericanEnglish, language.BritishEnglish, language.English:
		return englishFormatter
	default:
		return defaultFormatter
	}
}
