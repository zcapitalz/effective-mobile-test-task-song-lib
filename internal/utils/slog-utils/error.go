package slogutils

import (
	"context"
	"log/slog"
	"song-lib/internal/utils"
)

func Error(ctx context.Context, msg string, err error, args ...any) {
	args = append(args, slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	})
	utils.ContextLogger(ctx).Error(
		msg,
		args...,
	)
}
