package view

import (
	"context"
	"time"

	"github.com/dgf/go-ssr-x/locale"
)

func localizeDate(ctx context.Context, d time.Time) string {
	return locale.LocalizeDate(ctx, d)
}

func localizeDateTime(ctx context.Context, dt time.Time) string {
	return locale.LocalizeDateTime(ctx, dt)
}

func translate(ctx context.Context, messageID string) string {
	return locale.Translate(ctx, messageID)
}

func translateData(ctx context.Context, messageID string, data map[string]string) string {
	return locale.TranslateData(ctx, messageID, data)
}
