package utils

import (
	"context"
	"log/slog"
)

func ContextLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func PassContextLogger(oldCtx context.Context, newCtx context.Context) context.Context {
	return context.WithValue(newCtx, "logger", ContextLogger(oldCtx))
}
