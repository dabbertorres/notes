package telemetry

import (
	"context"
	"os"

	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/dabbertorres/notes/config"
)

func SetupMetrics(injector do.Injector) (*metric.MeterProvider, error) {
	ctx := do.MustInvoke[context.Context](injector)
	cfg := do.MustInvoke[*config.Config](injector).Telemetry.Metrics
	res := do.MustInvoke[*resource.Resource](injector)

	var opts []metric.Option

	for _, d := range cfg.Destinations {
		var (
			ex  metric.Exporter
			err error
		)
		switch d {
		case config.TelemetryStdout:
			ex, err = stdoutmetric.New()

		case config.TelemetryStderr:
			ex, err = stdoutmetric.New(stdoutmetric.WithWriter(os.Stderr))

		case config.TelemetryOTLPGRPC:
			// otherwise configured via environment variables as documented here:
			// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#pkg-overview
			ex, err = otlpmetricgrpc.New(ctx)

		case config.TelemetryOTLPHTTP:
			// otherwise configured via environment variables as documented here:
			// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#pkg-overview
			ex, err = otlpmetrichttp.New(ctx)
		}

		if err != nil {
			return nil, err
		}

		opts = append(opts, metric.WithReader(metric.NewPeriodicReader(ex)))
	}

	opts = append(opts, metric.WithResource(res))

	provider := metric.NewMeterProvider(opts...)
	return provider, nil
}
