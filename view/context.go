package view

import (
	"context"
	"log/slog"

	"github.com/dgf/go-ssr-x/locales"
)

type ViewContextKey string

var TranslatorKey ViewContextKey = "translator"

func Translate(ctx context.Context, messageID string) string {
	if translator, ok := ctx.Value(TranslatorKey).(locales.Translator); ok {
		return translator.Translate(messageID)
	}
	slog.Warn("view context contains no translator")
	return messageID
}

func TranslateData(ctx context.Context, messageID string, data map[string]string) string {
	if translator, ok := ctx.Value(TranslatorKey).(locales.Translator); ok {
		return translator.TranslateData(messageID, data)
	}
	slog.Warn("view context contains no translator")
	return messageID
}
