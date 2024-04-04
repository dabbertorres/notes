package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/samber/do"
)

const (
	// PathName is the DI identifier for the value containing the path to the config file.
	PathName = "config-path"
)

type Config struct {
	Database Database `json:"database"`
	HTTP     HTTP     `json:"http"`
	Logging  Logging  `json:"logging"`
}

func Load(injector *do.Injector) (*Config, error) {
	path, err := do.InvokeNamed[string](injector, PathName)
	if err != nil {
		return nil, err
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(buf, &cfg); err != nil {
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
	c.Logging.applyDefaults()
}

func (c *Config) validate() error {
	var errs []error

	if err := c.Database.validate(); err != nil {
		errs = append(errs, err)
	}

	if err := c.HTTP.validate(); err != nil {
		errs = append(errs, err)
	}

	if err := c.Logging.validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
