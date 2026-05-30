package config

type ESP32Config struct {
	DeviceURL string `env:"DEVICE_URL" validate:"required"`
}
