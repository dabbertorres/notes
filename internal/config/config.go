package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/samber/do/v2"

	"github.com/dabbertorres/notes/internal/util"
)

const (
	// PathName is the DI identifier for the value containing the path to the config file.
	PathName = "config-path"
)

type Config struct {
	Database  Database  `json:"database"`
	HTTP      HTTP      `json:"http"`
	Telemetry Telemetry `json:"telemetry"`
}

type decoder interface {
	Decode(any) error
}

func Load(injector do.Injector) (*Config, error) {
	path, err := do.InvokeNamed[string](injector, PathName)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var dec decoder

	switch ext := filepath.Ext(path); ext {
	case ".json":
		dec = util.Apply(json.NewDecoder(f), func(d *json.Decoder) {
			d.DisallowUnknownFields()
		})

	case ".yaml":
		dec = yaml.NewDecoder(f,
			yaml.DisallowDuplicateKey(),
			yaml.DisallowUnknownField(),
			yaml.UseJSONUnmarshaler(),
		)

	default:
		return nil, fmt.Errorf("config file is unsupported format: %q", ext)
	}

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	c.Database.applyDefaults()
	c.HTTP.applyDefaults()
	c.Telemetry.applyDefaults()
}

func (c *Config) validate() error {
	var errs fieldErrorList

	if err := c.Database.validate(); err != nil {
		errs = append(errs, err.qualify(".database")...)
	}

	if err := c.HTTP.validate(); err != nil {
		errs = append(errs, err.qualify(".http")...)
	}

	if err := c.Telemetry.validate(); err != nil {
		errs = append(errs, err.qualify(".telemetry")...)
	}

	return errors.Join(errs.asErrors()...)
}
