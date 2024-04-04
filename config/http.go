package config

import (
	"errors"
	"net/http"
	"time"
)

type HTTP struct {
	Addr              string        `json:"addr"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	MaxHeaderBytes    uint          `json:"max_header_bytes"`
	LogConnections    bool          `json:"log_connections"`
}

func (h *HTTP) applyDefaults() {
	if h.Addr == "" {
		h.Addr = ":8080"
	}

	if h.ReadHeaderTimeout == 0 {
		h.ReadHeaderTimeout = 2 * time.Second
	}

	if h.IdleTimeout == 0 {
		h.IdleTimeout = 5 * time.Minute
	}

	if h.MaxHeaderBytes == 0 {
		h.MaxHeaderBytes = http.DefaultMaxHeaderBytes
	}
}

func (h *HTTP) validate() error {
	var errs []error

	if h.Addr == "" {
		errs = append(errs, errors.New(".http.addr is requiredd"))
	}

	return errors.Join(errs...)
}
