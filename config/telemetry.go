package config

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type Telemetry struct {
	Logging Logging `json:"logging"`
	Tracing Tracing `json:"tracing"`
	Metrics Metrics `json:"metrics"`
}

func (t *Telemetry) applyDefaults() {
	t.Logging.applyDefaults()
	t.Tracing.applyDefaults()
	t.Metrics.applyDefaults()
}

func (t *Telemetry) validate() (errs fieldErrorList) {
	if err := t.Logging.validate(); err != nil {
		errs = append(errs, err.qualify(".logging")...)
	}

	if err := t.Tracing.validate(); err != nil {
		errs = append(errs, err.qualify(".tracing")...)
	}

	if err := t.Metrics.validate(); err != nil {
		errs = append(errs, err.qualify(".metrics")...)
	}

	return errs
}

type Logging struct {
	Destinations TelemetryDestinationList `json:"destinations"`
	Level        zapcore.Level            `json:"level"`
}

func (l *Logging) applyDefaults() {
	if len(l.Destinations) == 0 {
		l.Destinations = []TelemetryDestination{TelemetryStderr}
	}
}

func (l *Logging) validate() (errs fieldErrorList) {
	if err := l.Destinations.validate(); err != nil {
		errs = append(errs, err.qualify(".destinations")...)
	}

	return errs
}

type Tracing struct {
	Destinations TelemetryDestinationList `json:"destinations"`
}

func (t *Tracing) applyDefaults() {
	if len(t.Destinations) == 0 {
		t.Destinations = []TelemetryDestination{TelemetryStderr}
	}
}

func (t *Tracing) validate() (errs fieldErrorList) {
	if err := t.Destinations.validate(); err != nil {
		errs = append(errs, err.qualify(".destinations")...)
	}

	return errs
}

type Metrics struct {
	Destinations TelemetryDestinationList `json:"destinations"`
}

func (m *Metrics) applyDefaults() {
	if len(m.Destinations) == 0 {
		m.Destinations = []TelemetryDestination{TelemetryStderr}
	}
}

func (m *Metrics) validate() (errs fieldErrorList) {
	if err := m.Destinations.validate(); err != nil {
		errs = append(errs, err.qualify(".destinations")...)
	}

	return errs
}

type TelemetryDestination string

const (
	TelemetryStdout   TelemetryDestination = "stdout"
	TelemetryStderr   TelemetryDestination = "stderr"
	TelemetryOTLPGRPC TelemetryDestination = "otlp-grpc"
	TelemetryOTLPHTTP TelemetryDestination = "otlp-http"
)

func (t TelemetryDestination) validate() error {
	switch t {
	case TelemetryStdout,
		TelemetryStderr,
		TelemetryOTLPGRPC,
		TelemetryOTLPHTTP:
		return nil

	default:
		return fmt.Errorf("invalid value; must be one of %q, %q, %q, %q",
			TelemetryStdout,
			TelemetryStderr,
			TelemetryOTLPGRPC,
			TelemetryOTLPHTTP,
		)
	}
}

type TelemetryDestinationList []TelemetryDestination

func (l TelemetryDestinationList) validate() (errs fieldErrorList) {
	for i, d := range l {
		if err := d.validate(); err != nil {
			errs = append(errs, fieldError{fmt.Sprintf("[%d]", i), err.Error()})
		}
	}

	return errs
}
