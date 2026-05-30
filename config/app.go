package config

import "time"

type AppConfig struct {
	FetchInterval time.Duration `env:"FETCH_INTERVAL" envDefault:"60s"`
}
