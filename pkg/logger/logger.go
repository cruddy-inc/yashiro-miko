package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

type key struct{}

func NewLogger(serviceName string) *slog.Logger {
	logger := otelslog.NewLogger(serviceName)
	slog.SetDefault(logger)
	return logger
}

func SetLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, log)
}

func GetLogger(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(key{}).(*slog.Logger)
	if ok {
		return log
	}
	return slog.Default()
}
