package config

type OTelConfig struct {
	Endpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" validate:"required"`
	Insecure bool   `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"false"`
}
