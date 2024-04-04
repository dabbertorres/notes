package config

import (
	"errors"
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
	ConnectTimeout        time.Duration     `json:"connect_timeout"`
	MaxConnLifetime       time.Duration     `json:"max_conn_lifetime"`
	MaxConnLifetimeJitter time.Duration     `json:"max_conn_lifetime_jitter"`
	MaxConnIdleTime       time.Duration     `json:"max_conn_idle_time"`
	MaxConns              int               `json:"max_conns"`
	MinConns              int               `json:"min_conns"`
	HealthCheckPeriod     time.Duration     `json:"health_check_period"`
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

	if d.ConnectTimeout == 0 {
		d.ConnectTimeout = 2 * time.Minute
	}

	if d.MaxConnLifetime <= 0 {
		d.MaxConnLifetime = 1 * time.Hour
	}

	if d.MaxConnLifetimeJitter <= 0 {
		d.MaxConnLifetimeJitter = 1 * time.Minute
	}

	if d.MaxConnIdleTime <= 0 {
		d.MaxConnIdleTime = 15 * time.Minute
	}

	if d.MaxConns <= 0 {
		d.MaxConns = max(4, runtime.NumCPU())
	}

	if d.MinConns <= 0 {
		d.MinConns = 1
	}

	if d.HealthCheckPeriod == 0 {
		d.HealthCheckPeriod = 1 * time.Minute
	}
}

func (d *Database) validate() error {
	var errs []error

	if d.Host == "" {
		errs = append(errs, errors.New(".database.host is required"))
	}

	if d.Port == 0 {
		errs = append(errs, errors.New(".database.port is required"))
	}

	if d.User == "" {
		errs = append(errs, errors.New(".database.user is required"))
	}

	if d.Pass == "" {
		errs = append(errs, errors.New(".database.pass is required"))
	}

	if d.Name == "" {
		errs = append(errs, errors.New(".database.name is required"))
	}

	return errors.Join(errs...)
}
