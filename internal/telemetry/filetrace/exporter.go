package filetrace

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/embedded"
)

var ErrShutdown = errors.New("exporter is already shutdown")

type Exporter struct {
	embedded.TracerProvider
	f          *os.File
	enc        *json.Encoder
	mu         sync.Mutex
	isShutdown atomic.Bool
}

func New(f *os.File, opts ...Option) *Exporter {
	ex := &Exporter{
		f:   f,
		enc: json.NewEncoder(f),
	}

	for _, opt := range opts {
		opt(ex)
	}

	return ex
}

type Option func(*Exporter)

func WithPrettyPrint() Option {
	return func(e *Exporter) {
		e.enc.SetIndent("", "\t")
	}
}

func (e *Exporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if e.isShutdown.Load() {
		return ErrShutdown
	}

	if len(spans) == 0 {
		return nil
	}

	stubs := tracetest.SpanStubsFromReadOnlySpans(spans)

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, span := range stubs {
		if err := e.enc.Encode(span); err != nil {
			return err
		}
	}

	return nil
}

func (e *Exporter) Shutdown(ctx context.Context) error {
	if e.isShutdown.CompareAndSwap(false, true) {
		e.mu.Lock()
		defer e.mu.Unlock()

		return e.f.Close()
	}

	return ErrShutdown
}
