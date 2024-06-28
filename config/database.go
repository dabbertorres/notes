package config

import (
	"runtime"
	"time"
)

type Database struct {
	Host                  string            `json:"host"`
	Port                  uint16            `json:"port"`
	User                  string            `json:"user"`
	Pass                  string            `json:"pass"`
	Name                  string            `json:"name"`
	Args                  map[string]string `json:"args"`
	ConnectTimeout        Duration          `json:"connect_timeout"`
	MaxConnLifetime       Duration          `json:"max_conn_lifetime"`
	MaxConnLifetimeJitter Duration          `json:"max_conn_lifetime_jitter"`
	MaxConnIdleTime       Duration          `json:"max_conn_idle_time"`
	MaxConns              int               `json:"max_conns"`
	MinConns              int               `json:"min_conns"`
	HealthCheckPeriod     Duration          `json:"health_check_period"`
	LogConnections        bool              `json:"log_connections"`
}

func (d *Database) applyDefaults() {
	if d.Host == "" {
		d.Host = "localhost"
	}

	if d.Port == 0 {
		d.Port = 5432
	}

	if d.User == "" {
		d.User = "postgres"
	}

	if d.Pass == "" {
		d.Pass = "postgres"
	}

	if d.Name == "" {
		d.Name = "postgres"
	}

	if d.Args == nil {
		d.Args = make(map[string]string)
	}

	if d.ConnectTimeout.Value == 0 {
		d.ConnectTimeout.Value = 2 * time.Minute
	}

	if d.MaxConnLifetime.Value <= 0 {
		d.MaxConnLifetime.Value = 1 * time.Hour
	}

	if d.MaxConnLifetimeJitter.Value <= 0 {
		d.MaxConnLifetimeJitter.Value = 1 * time.Minute
	}

	if d.MaxConnIdleTime.Value <= 0 {
		d.MaxConnIdleTime.Value = 15 * time.Minute
	}

	if d.MaxConns <= 0 {
		d.MaxConns = max(4, runtime.NumCPU())
	}

	if d.MinConns <= 0 {
		d.MinConns = 1
	}

	if d.HealthCheckPeriod.Value == 0 {
		d.HealthCheckPeriod.Value = 1 * time.Minute
	}
}

func (d *Database) validate() (errs fieldErrorList) {
	if d.Host == "" {
		errs = append(errs, fieldError{".host", "is required"})
	}

	if d.Port == 0 {
		errs = append(errs, fieldError{".port", "is required"})
	}

	if d.User == "" {
		errs = append(errs, fieldError{".user", "is required"})
	}

	if d.Pass == "" {
		errs = append(errs, fieldError{".pass", "is required"})
	}

	if d.Name == "" {
		errs = append(errs, fieldError{".name", "is required"})
	}

	return errs
}
