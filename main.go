package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/walnuts1018/esp32-thermohygrometer-exporter/config"
	"github.com/walnuts1018/esp32-thermohygrometer-exporter/esp32"
	"github.com/walnuts1018/esp32-thermohygrometer-exporter/logger"
	"github.com/walnuts1018/esp32-thermohygrometer-exporter/otel"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, os.Kill)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger.CreateLogger(cfg.LogLevel, cfg.LogType))

	exporter, shutdownOtel, err := otel.NewExporter(ctx, cfg)
	if err != nil {
		slog.Error("failed to initialize OTel exporter", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdownOtel(context.Background()); err != nil {
			slog.Error("failed to shutdown OpenTelemetry", "error", err)
		}
	}()

	client := esp32.NewClient(ctx, cfg)

	slog.Info("Starting ESP32 Thermohygrometer Exporter", "interval", cfg.App.FetchInterval)

	ticker := time.NewTicker(cfg.App.FetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down periodic export")
			return
		case <-ticker.C:
			fetchAndExport(ctx, client, exporter)
		}
	}
}

func fetchAndExport(ctx context.Context, client *esp32.Client, exporter *otel.Exporter) {
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	m, err := client.FetchLatest(fetchCtx)
	if err != nil {
		slog.Error("failed to fetch latest measurement", "error", err)
		return
	}

	if err := exporter.Export(ctx, m); err != nil {
		slog.Error("failed to export measurement", "error", err)
		return
	}

	slog.Info("successfully exported measurement",
		"temperature_celsius", m.TemperatureCelsius,
		"relative_humidity_percent", m.RelativeHumidityPercent,
		"sensor", m.Sensor,
	)
}
