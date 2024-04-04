package config

import (
	"go.uber.org/zap/zapcore"
)

type Logging struct {
	Level zapcore.Level `json:"level"`
}

func (l *Logging) applyDefaults() {}

func (l *Logging) validate() error { return nil }
