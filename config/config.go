package config

import (
	"log/slog"
	"reflect"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	LogLevel slog.Level `env:"LOG_LEVEL" envDefault:"info"`
	LogType  LogType    `env:"LOG_TYPE" envDefault:"json"`
	App      AppConfig
	ESP32    ESP32Config
	OIDC     OIDCConfig
	OTel     OTelConfig
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.ParseWithOptions(&cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeFor[slog.Level](): returnAny(ParseLogLevel),
			reflect.TypeFor[LogType]():    returnAny(ParseLogType),
		},
	}); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func returnAny[T any](f func(v string) (t T, err error)) env.ParserFunc {
	return func(v string) (any, error) {
		t, err := f(v)
		return any(t), err
	}
}
