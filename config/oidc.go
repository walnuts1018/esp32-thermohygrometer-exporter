package config

type OIDCConfig struct {
	TokenURL     string   `env:"OIDC_TOKEN_URL" validate:"required"`
	ClientID     string   `env:"OIDC_CLIENT_ID" validate:"required"`
	ClientSecret string   `env:"OIDC_CLIENT_SECRET" validate:"required"`
	Scopes       []string `env:"OIDC_SCOPES" envSeparator:" "`
	Audience     string   `env:"OIDC_AUDIENCE"`
}
