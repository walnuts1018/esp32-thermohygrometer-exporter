# esp32-thermohygrometer-exporter

[![CI](https://github.com/walnuts1018/esp32-thermohygrometer-exporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/walnuts1018/esp32-thermohygrometer-exporter/actions/workflows/ci.yaml)
[![Docker](https://github.com/walnuts1018/esp32-thermohygrometer-exporter/actions/workflows/docker.yaml/badge.svg)](https://github.com/walnuts1018/esp32-thermohygrometer-exporter/actions/workflows/docker.yaml)

`esp32-thermohygrometer-exporter` is an application that fetches temperature and humidity from [esp32-thermohygrometer](https://github.com/walnuts1018/esp32-thermohygrometer) and continuously exports these measurements as OpenTelemetry metrics.

## Features

- **Secure Data Fetching**: Authenticates with OIDC (Client Credentials flow) to securely access the ESP32 endpoint.
- **OpenTelemetry Metrics**: Exports the retrieved temperature (`esp32.temperature`) and humidity (`esp32.humidity`) as OpenTelemetry gauges.
- **Configurable Polling**: Fetches new measurements on a configurable interval.
- **Clean Architecture**: Follows Domain-Driven Design (DDD) principles for maintainability.

## Configuration

Configuration is loaded from environment variables using `caarlos0/env` and validated using `validator/v10`.

| Name | Description | Default |
| --- | --- | --- |
| `FETCH_INTERVAL` | Polling interval for measurements (e.g. `60s`, `1m`) | `60s` |
| `DEVICE_URL` | Base URL of the ESP32 device endpoint | **Required** |
| `OIDC_TOKEN_URL` | OIDC token URL | **Required** |
| `OIDC_PRIVATE_KEY_JSON`| ZITADEL Service Account JSON private key content | **Required** |
| `OIDC_SCOPES` | Space-separated list of scopes to request | |
| `OTEL_EXPORTER_OTLP_ENDPOINT`| OTLP gRPC endpoint for the collector | **Required** |
| `OTEL_EXPORTER_OTLP_INSECURE`| Whether to connect to OTLP collector without TLS | `false` |
| `LOG_LEVEL` | Application logging level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_TYPE` | Output format for logs (`text`, `json`) | `json` |

## Development

### Running Locally

To run the application locally (using `mise`):

```bash
mise run build
mise run run
```

### Testing and Linting

```bash
mise run lint
mise run test
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
