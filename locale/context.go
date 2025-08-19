// Package locale provides a combined i18n and l18n context.
package locale

import (
	"context"
	"time"

	"golang.org/x/text/language"
)

type ContextKey string

const LocaleContextKey ContextKey = "locale"

type localeContext struct {
	*formatter
	*translator

	lang language.Tag
}

func newContext(lang language.Tag) *localeContext {
	return &localeContext{
		lang:       lang,
		formatter:  refFormatter(lang),
		translator: refTranslator(lang),
	}
}

func WithLocale(ctx context.Context, lang language.Tag) context.Context {
	return context.WithValue(ctx, LocaleContextKey, newContext(lang))
}

func LocalizeDate(ctx context.Context, d time.Time) string {
	l, ok := ctx.Value(LocaleContextKey).(*localeContext)
	if !ok {
		return d.Format(time.DateOnly)
	}

	return l.formatDate(d)
}

func LocalizeDateTime(ctx context.Context, dt time.Time) string {
	l, ok := ctx.Value(LocaleContextKey).(*localeContext)
	if !ok {
		return dt.Format(time.DateTime)
	}

	return l.formatDateTime(dt)
}

func Translate(ctx context.Context, messageID string) string {
	l, ok := ctx.Value(LocaleContextKey).(*localeContext)
	if !ok {
		return messageID
	}

	return l.translate(messageID)
}

func TranslateData(ctx context.Context, messageID string, data map[string]string) string {
	l, ok := ctx.Value(LocaleContextKey).(*localeContext)
	if !ok {
		return messageID
	}

	return l.translateData(messageID, data)
}
