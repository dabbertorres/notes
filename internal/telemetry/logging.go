package telemetry

import (
	"context"
	"os"

	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/dabbertorres/notes/internal/config"
)

func SetupLogging(injector do.Injector) (*log.LoggerProvider, error) {
	ctx := do.MustInvoke[context.Context](injector)
	cfg := do.MustInvoke[*config.Config](injector).Telemetry.Logging
	res := do.MustInvoke[*resource.Resource](injector)

	var opts []log.LoggerProviderOption

	for _, d := range cfg.Destinations {
		var (
			ex  log.Exporter
			err error
		)
		switch d {
		case config.TelemetryStdout:
			ex, err = stdoutlog.New()

		case config.TelemetryStderr:
			ex, err = stdoutlog.New(stdoutlog.WithWriter(os.Stderr))

		case config.TelemetryOTLPGRPC:
			// otherwise configured via environment variables as documented here:
			// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc#pkg-overview
			ex, err = otlploggrpc.New(ctx)

		case config.TelemetryOTLPHTTP:
			// otherwise configured via environment variables as documented here:
			// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp#pkg-overview
			ex, err = otlploghttp.New(ctx)
		}

		if err != nil {
			return nil, err
		}

		opts = append(opts, log.WithProcessor(log.NewBatchProcessor(ex)))
	}

	opts = append(opts, log.WithResource(res))

	provider := log.NewLoggerProvider(opts...)
	return provider, nil
}
