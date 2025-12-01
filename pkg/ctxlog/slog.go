package ctxlog

import (
	"context"
	"log/slog"
	"os"
)

func NewLogger(ctx context.Context) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	return logger
}

var ctxKeyLogger = struct{}{}

func SetLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger := ctx.Value(ctxKeyLogger).(*slog.Logger)
	if logger != nil {
		return logger
	}

	return NewLogger(ctx)
}
