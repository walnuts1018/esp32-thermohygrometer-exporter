package otel

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/walnuts1018/esp32-thermohygrometer-exporter/config"
	"github.com/walnuts1018/esp32-thermohygrometer-exporter/esp32"
)

type Exporter struct {
	meterProvider *sdkmetric.MeterProvider

	mu          sync.Mutex
	latestTemp  float64
	latestHum   float64
	latestAttrs []attribute.KeyValue
	hasData     bool
}

func NewExporter(ctx context.Context, cfg *config.Config) (*Exporter, func(context.Context) error, error) {
	endpoint := cfg.OTel.Endpoint
	if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
		endpoint = u.Host
	}

	exporterOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
	}
	if cfg.OTel.Insecure {
		exporterOpts = append(exporterOpts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create otlp metric exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("esp32-thermohygrometer-exporter"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(cfg.App.FetchInterval))),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(meterProvider)

	e := &Exporter{
		meterProvider: meterProvider,
	}

	meter := meterProvider.Meter("esp32-thermohygrometer-exporter")

	tempGauge, err := meter.Float64ObservableGauge("esp32.temperature", metric.WithDescription("Temperature in Celsius"), metric.WithUnit("Cel"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temperature gauge: %w", err)
	}

	humGauge, err := meter.Float64ObservableGauge("esp32.humidity", metric.WithDescription("Relative humidity in percent"), metric.WithUnit("%"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create humidity gauge: %w", err)
	}

	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		e.mu.Lock()
		defer e.mu.Unlock()

		if e.hasData {
			o.ObserveFloat64(tempGauge, e.latestTemp, metric.WithAttributes(e.latestAttrs...))
			o.ObserveFloat64(humGauge, e.latestHum, metric.WithAttributes(e.latestAttrs...))
		}
		return nil
	}, tempGauge, humGauge)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register callback: %w", err)
	}

	return e, meterProvider.Shutdown, nil
}

func (e *Exporter) Export(ctx context.Context, m *esp32.Measurement) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.latestTemp = m.TemperatureCelsius
	e.latestHum = m.RelativeHumidityPercent
	e.latestAttrs = []attribute.KeyValue{
		attribute.String("sensor", m.Sensor),
	}
	e.hasData = true

	return nil
}
