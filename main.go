package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
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

	client, err := esp32.NewClient(ctx, cfg)
	if err != nil {
		slog.Error("failed to create esp32 client", "error", err)
		os.Exit(1)
	}

	var isReady atomic.Bool

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if isReady.Load() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.ProbePort),
		Handler: mux,
	}

	go func() {
		slog.Info("Starting probe server", "port", cfg.App.ProbePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("probe server failed", "error", err)
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("Starting ESP32 Thermohygrometer Exporter", "interval", cfg.App.FetchInterval)

	ticker := time.NewTicker(cfg.App.FetchInterval)
	defer ticker.Stop()

	if err := fetchAndExport(ctx, client, exporter); err == nil {
		isReady.Store(true)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down periodic export")
			return
		case <-ticker.C:
			if err := fetchAndExport(ctx, client, exporter); err == nil {
				isReady.Store(true)
			}
		}
	}
}

func fetchAndExport(ctx context.Context, client *esp32.Client, exporter *otel.Exporter) error {
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	m, err := client.FetchLatest(fetchCtx)
	if err != nil {
		slog.Error("failed to fetch latest measurement", "error", err)
		return err
	}

	exporter.Export(ctx, m)

	slog.Info("successfully exported measurement",
		"temperature_celsius", m.TemperatureCelsius,
		"relative_humidity_percent", m.RelativeHumidityPercent,
		"sensor", m.Sensor,
	)
	return nil
}
