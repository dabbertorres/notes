package telemetry

import (
	"context"

	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
)

var Package = do.Package(
	do.Lazy(DetectResource),
	do.Lazy(SetupTracing),
	do.Lazy(SetupMetrics),
	do.Lazy(SetupLogging),
	do.Eager(Setup),
)

type Service struct{}

func Setup(injector do.Injector) (svc Service, err error) {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	tracer, err := do.InvokeAs[trace.TracerProvider](injector)
	if err != nil {
		return svc, err
	}

	meter, err := do.InvokeAs[metric.MeterProvider](injector)
	if err != nil {
		return svc, err
	}

	logger, err := do.InvokeAs[log.LoggerProvider](injector)
	if err != nil {
		return svc, err
	}

	otel.SetTextMapPropagator(propagator)
	// otel.SetErrorHandler(svc)

	otel.SetTracerProvider(tracer)
	otel.SetMeterProvider(meter)
	global.SetLoggerProvider(logger)

	return svc, nil
}

func (s Service) Handle(err error) {}

func DetectResource(injector do.Injector) (*resource.Resource, error) {
	ctx := do.MustInvoke[context.Context](injector)

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("notes"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
