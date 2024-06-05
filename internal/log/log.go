package log

import (
	"context"

	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/scope"
)

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).Debug(msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).Error(msg, fields...)
}

func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).DPanic(msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).Panic(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).Fatal(msg, fields...)
}
