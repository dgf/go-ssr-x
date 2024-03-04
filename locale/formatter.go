package locale

import (
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/text/language"
)

type Formatter interface {
	FormatDate(d time.Time) string
	FormatDateTime(dt time.Time) string
}

type formatter struct {
	formatDate     func(d time.Time) string
	formatDateTime func(dt time.Time) string
}

func acceptOrDefault(r *http.Request) language.Tag {
	accept := r.Header.Get("Accept-Language")
	if tags, _, err := language.ParseAcceptLanguage(accept); err != nil {
		slog.Info("accept language header parse failed", "header", accept)
		return language.English
	} else {
		return tags[0]
	}
}

func NewFormatter(r *http.Request) Formatter {
	lang := acceptOrDefault(r)
	switch lang {
	case language.German:
		return &formatter{
			formatDate: func(d time.Time) string {
				return d.Format("02.01.2006")
			},
			formatDateTime: func(dt time.Time) string {
				return dt.Format("02.01.2006 15:04:05")
			},
		}
	case language.AmericanEnglish, language.BritishEnglish, language.English:
		return &formatter{
			formatDate: func(d time.Time) string {
				return d.Format("01/02/2006")
			},
			formatDateTime: func(dt time.Time) string {
				return dt.Format("01/02/2006 3:04PM")
			},
		}
	default:
		return &formatter{
			formatDate: func(d time.Time) string {
				return d.Format(time.DateOnly)
			},
			formatDateTime: func(dt time.Time) string {
				return dt.Format(time.DateTime)
			},
		}
	}
}

func (f *formatter) FormatDate(d time.Time) string {
	return f.formatDate(d)
}

func (f *formatter) FormatDateTime(dt time.Time) string {
	return f.formatDateTime(dt)
}
