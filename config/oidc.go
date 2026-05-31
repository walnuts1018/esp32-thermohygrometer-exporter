package config

type OIDCConfig struct {
	TokenURL       string   `env:"OIDC_TOKEN_URL" validate:"required"`
	JSONKeyContent string   `env:"OIDC_PRIVATE_KEY_JSON" validate:"required"`
	Scopes         []string `env:"OIDC_SCOPES" envSeparator:" "`
	Audience       string   `env:"OIDC_AUDIENCE"`
}
