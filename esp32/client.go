package esp32

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"golang.org/x/oauth2"

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

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	keyFile, err := client.ConfigFromKeyFileData([]byte(cfg.OIDC.JSONKeyContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ZITADEL key data: %w", err)
	}

	initializer := client.AuthenticationJWTProfile(keyFile, cfg.OIDC.Scopes...)
	tokenSource, err := initializer(ctx, cfg.OIDC.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %w", err)
	}

	oauthClient := oauth2.NewClient(ctx, tokenSource)

	return &Client{
		deviceURL: cfg.ESP32.DeviceURL,
		client:    oauthClient,
	}, nil
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
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			fmt.Printf("failed to read response body: %v\n", err)
		}
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
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
