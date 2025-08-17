// Package log wraps the concrete logger implementation (currently slog).
package log

import (
	"log/slog"
	"os"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

func Error(msg string, err error) {
	slog.Error(msg, slog.String("error", err.Error()))
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}
