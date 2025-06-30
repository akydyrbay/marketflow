package logger

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init() {
	// TextHandler or JSONHandler, whichever you like:
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // optional: include file/line in output
	})
	Log = slog.New(handler)
}

func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

func WithContext(ctx context.Context) *slog.Logger {
	return slog.Default().With(slog.Any("ctx", ctx))
}
