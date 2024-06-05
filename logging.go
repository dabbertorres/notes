package main

import (
	"os"
	"time"

	"github.com/samber/do"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dabbertorres/notes/config"
)

func setupLogging(injector *do.Injector) (*zap.Logger, error) {
	logLevel := zap.NewAtomicLevel()

	cfg := do.MustInvoke[*config.Config](injector).Logging

	logLevel.SetLevel(cfg.Level)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "name",
			CallerKey:      "caller",
			FunctionKey:    "func",
			StacktraceKey:  "stack",
			SkipLineEnding: false,
			LineEnding:     "\n",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format(time.RFC3339Nano))
			},
			EncodeDuration:   zapcore.MillisDurationEncoder,
			EncodeCaller:     zapcore.FullCallerEncoder,
			EncodeName:       zapcore.FullNameEncoder,
			ConsoleSeparator: "",
		}),
		os.Stderr,
		logLevel,
	)

	logger := zap.New(
		core,
		zap.WithCaller(true),
	)

	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger)

	return logger, nil
}
