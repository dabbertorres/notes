package config

import "time"

type Duration struct {
	Value time.Duration
}

func (d *Duration) UnmarshalText(data []byte) error {
	v, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}

	d.Value = v
	return nil
}
