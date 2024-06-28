package config

import (
	"net/http"
	"time"
)

type HTTP struct {
	Addr              string   `json:"addr"`
	ReadHeaderTimeout Duration `json:"read_header_timeout"`
	IdleTimeout       Duration `json:"idle_timeout"`
	MaxHeaderBytes    uint     `json:"max_header_bytes"`
	LogConnections    bool     `json:"log_connections"`
}

func (h *HTTP) applyDefaults() {
	if h.Addr == "" {
		h.Addr = ":8080"
	}

	if h.ReadHeaderTimeout.Value == 0 {
		h.ReadHeaderTimeout.Value = 2 * time.Second
	}

	if h.IdleTimeout.Value == 0 {
		h.IdleTimeout.Value = 5 * time.Minute
	}

	if h.MaxHeaderBytes == 0 {
		h.MaxHeaderBytes = http.DefaultMaxHeaderBytes
	}
}

func (h *HTTP) validate() (errs fieldErrorList) {
	if h.Addr == "" {
		errs = append(errs, fieldError{".addr", "is required"})
	}

	return errs
}
