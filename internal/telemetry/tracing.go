package telemetry

import (
	"context"
	"os"

	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/dabbertorres/notes/config"
)

func SetupTracing(injector do.Injector) (*trace.TracerProvider, error) {
	ctx := do.MustInvoke[context.Context](injector)
	cfg := do.MustInvoke[*config.Config](injector).Telemetry.Tracing
	res := do.MustInvoke[*resource.Resource](injector)

	var opts []trace.TracerProviderOption

	for _, d := range cfg.Destinations {
		var (
			ex  trace.SpanExporter
			err error
		)
		switch d {
		case config.TelemetryStdout:
			ex, err = stdouttrace.New()

		case config.TelemetryStderr:
			ex, err = stdouttrace.New(stdouttrace.WithWriter(os.Stderr))

		case config.TelemetryOTLPGRPC:
			// otherwise configured via environment variables as documented here:
			// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc#pkg-overview
			ex, err = otlptracegrpc.New(ctx)

		case config.TelemetryOTLPHTTP:
			ex, err = otlptracehttp.New(ctx)
		}

		if err != nil {
			return nil, err
		}

		opts = append(opts, trace.WithBatcher(ex))
	}

	opts = append(opts, trace.WithResource(res))

	provider := trace.NewTracerProvider(opts...)
	return provider, nil
}
