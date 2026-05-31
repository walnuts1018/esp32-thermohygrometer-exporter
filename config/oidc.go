package config

type OIDCConfig struct {
	Issuer         string   `env:"OIDC_ISSUER" validate:"required"`
	JSONKeyContent string   `env:"OIDC_PRIVATE_KEY_JSON" validate:"required"`
	Scopes         []string `env:"OIDC_SCOPES" envSeparator:" "`
}
