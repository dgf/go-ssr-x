package view

import (
	"context"
	"log/slog"
	"time"

	"github.com/dgf/go-ssr-x/locale"
)

type ViewContextKey string

var LocaleHelperKey ViewContextKey = "locale"

type LocaleHelper struct {
	Formatter  locale.Formatter
	Translator locale.Translator
}

func LocalizeDate(ctx context.Context, d time.Time) string {
	if h, ok := ctx.Value(LocaleHelperKey).(LocaleHelper); ok {
		return h.Formatter.FormatDate(d)
	}
	slog.Warn("view context contains no fromatter")
	return d.Format(time.DateOnly)
}

func LocalizeDateTime(ctx context.Context, dt time.Time) string {
	if h, ok := ctx.Value(LocaleHelperKey).(LocaleHelper); ok {
		return h.Formatter.FormatDateTime(dt)
	}
	slog.Warn("view context contains no fromatter")
	return dt.Format(time.DateTime)
}

func Translate(ctx context.Context, messageID string) string {
	if h, ok := ctx.Value(LocaleHelperKey).(LocaleHelper); ok {
		return h.Translator.Translate(messageID)
	}
	slog.Warn("view context contains no translator")
	return messageID
}

func TranslateData(ctx context.Context, messageID string, data map[string]string) string {
	if h, ok := ctx.Value(LocaleHelperKey).(LocaleHelper); ok {
		return h.Translator.TranslateData(messageID, data)
	}
	slog.Warn("view context contains no translator")
	return messageID
}
