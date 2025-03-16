package view

import (
	"context"
	"time"

	"github.com/dgf/go-ssr-x/locale"
	"github.com/dgf/go-ssr-x/log"
)

type ViewContextKey string

var LocaleContextKey ViewContextKey = "locale"

type LocaleContext struct {
	Formatter  locale.Formatter
	Translator locale.Translator
}

func runLocalized(ctx context.Context, localize func(LocaleContext) string, fallback func() string) string {
	if l, ok := ctx.Value(LocaleContextKey).(LocaleContext); ok {
		return localize(l)
	}
	log.Warn("view context contains no locale")
	return fallback()
}

func LocalizeDate(ctx context.Context, d time.Time) string {
	return runLocalized(ctx, func(l LocaleContext) string {
		return l.Formatter.FormatDate(d)
	}, func() string {
		return d.Format(time.DateOnly)
	})
}

func LocalizeDateTime(ctx context.Context, dt time.Time) string {
	return runLocalized(ctx, func(l LocaleContext) string {
		return l.Formatter.FormatDateTime(dt)
	}, func() string {
		return dt.Format(time.DateTime)
	})
}

func Translate(ctx context.Context, messageID string) string {
	return runLocalized(ctx, func(l LocaleContext) string {
		return l.Translator.Translate(messageID)
	}, func() string {
		return messageID
	})
}

func TranslateData(ctx context.Context, messageID string, data map[string]string) string {
	return runLocalized(ctx, func(l LocaleContext) string {
		return l.Translator.TranslateData(messageID, data)
	}, func() string {
		return messageID
	})
}
