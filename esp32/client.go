package esp32

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/walnuts1018/esp32-thermohygrometer-exporter/config"
)

type Measurement struct {
	TemperatureCelsius      float64
	RelativeHumidityPercent float64
	Sensor                  string
	MeasuredAtMS            int64
}

type Client struct {
	deviceURL string
	client    *http.Client
}

type esp32Measurement struct {
	TemperatureCelsius      float64 `json:"temperature_celsius"`
	RelativeHumidityPercent float64 `json:"relative_humidity_percent"`
	Sensor                  string  `json:"sensor"`
	I2CAddress              string  `json:"i2c_address"`
	MeasuredAtMS            int64   `json:"measured_at_ms"`
}

func NewClient(ctx context.Context, cfg *config.Config) *Client {
	params := url.Values{}
	if cfg.OIDC.Audience != "" {
		params.Set("audience", cfg.OIDC.Audience)
	}

	oauthConfig := clientcredentials.Config{
		ClientID:       cfg.OIDC.ClientID,
		ClientSecret:   cfg.OIDC.ClientSecret,
		TokenURL:       cfg.OIDC.TokenURL,
		Scopes:         cfg.OIDC.Scopes,
		EndpointParams: params,
		AuthStyle:      oauth2.AuthStyleInHeader,
	}

	client := oauth2.NewClient(ctx, oauthConfig.TokenSource(ctx))

	return &Client{
		deviceURL: cfg.ESP32.DeviceURL,
		client:    client,
	}
}

func (c *Client) FetchLatest(ctx context.Context) (*Measurement, error) {
	measureURL, err := url.JoinPath(c.deviceURL, "/v1/measurements/latest")
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, measureURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call measurement API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var latest esp32Measurement
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Measurement{
		TemperatureCelsius:      latest.TemperatureCelsius,
		RelativeHumidityPercent: latest.RelativeHumidityPercent,
		Sensor:                  latest.Sensor,
		MeasuredAtMS:            latest.MeasuredAtMS,
	}, nil
}
