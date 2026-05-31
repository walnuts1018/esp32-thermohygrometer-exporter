package config

import "time"

type AppConfig struct {
	FetchInterval time.Duration `env:"FETCH_INTERVAL" envDefault:"60s"`
	ProbePort     int           `env:"PROBE_PORT" envDefault:"8080"`
}
