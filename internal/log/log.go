package log

import (
	"context"

	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/scope"
)

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}

func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		DPanic(msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		Panic(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	scope.Logger(ctx).
		WithOptions(zap.AddCallerSkip(1)).
		Fatal(msg, fields...)
}
