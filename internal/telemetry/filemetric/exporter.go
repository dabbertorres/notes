package filemetric

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var ErrShutdown = errors.New("exporter is already shutdown")

type Exporter struct {
	embedded.MeterProvider
	f                   *os.File
	enc                 *json.Encoder
	mu                  sync.Mutex
	temporalitySelector metric.TemporalitySelector
	aggregationSelector metric.AggregationSelector
	isShutdown          atomic.Bool
}

func New(f *os.File, opts ...Option) *Exporter {
	ex := &Exporter{
		f:                   f,
		enc:                 json.NewEncoder(f),
		temporalitySelector: metric.DefaultTemporalitySelector,
		aggregationSelector: metric.DefaultAggregationSelector,
	}

	for _, opt := range opts {
		opt(ex)
	}

	return ex
}

type Option func(*Exporter)

func WithTemporalitySelector(selector metric.TemporalitySelector) Option {
	return func(e *Exporter) {
		e.temporalitySelector = selector
	}
}

func WithAggregationSelector(selector metric.AggregationSelector) Option {
	return func(e *Exporter) {
		e.aggregationSelector = selector
	}
}

func WithPrettyPrint() Option {
	return func(e *Exporter) {
		e.enc.SetIndent("", "\t")
	}
}

func (e *Exporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return e.temporalitySelector(kind)
}

func (e *Exporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return e.aggregationSelector(kind)
}

func (e *Exporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if e.isShutdown.Load() {
		return ErrShutdown
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enc.Encode(data)
}

func (e *Exporter) ForceFlush(ctx context.Context) error {
	if e.isShutdown.Load() {
		return ErrShutdown
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	return e.f.Sync()
}

func (e *Exporter) Shutdown(ctx context.Context) error {
	if e.isShutdown.CompareAndSwap(false, true) {
		e.mu.Lock()
		defer e.mu.Unlock()

		return e.f.Close()
	}

	return ErrShutdown
}
