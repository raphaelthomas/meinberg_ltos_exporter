# Prometheus Exporter for Meinberg LTOS

[![Go Version](https://img.shields.io/github/go-mod/go-version/raphaelthomas/meinberg_ltos_exporter)](https://github.com/raphaelthomas/meinberg_ltos_exporter/blob/main/go.mod)
[![Latest Release](https://img.shields.io/github/v/release/raphaelthomas/meinberg_ltos_exporter)](https://github.com/raphaelthomas/meinberg_ltos_exporter/releases/latest)
[![License](https://img.shields.io/github/license/raphaelthomas/meinberg_ltos_exporter)](https://github.com/raphaelthomas/meinberg_ltos_exporter/blob/main/LICENSE)
[![CI](https://github.com/raphaelthomas/meinberg_ltos_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/raphaelthomas/meinberg_ltos_exporter/actions/workflows/ci.yml)

Prometheus exporter `meinberg_ltos_exporter` is designed for Meinberg devices
running LTOS. It retrieves the status of a device via its REST API and exposes
the available data as Prometheus metrics.

**New here?** Check out the [Getting Started guide](GETTING_STARTED.md).

> [!IMPORTANT]
> This exporter is **experimental** and has only been tested against a very
> limited number of devices. It may not work correctly with all Meinberg
> devices and LTOS versions.
>
> Please send a **[device
> report](https://github.com/raphaelthomas/meinberg_ltos_exporter/issues/new?template=device-report.yml)**
> for any bugs and to help extend compatibility. Include the anonymized JSON
> output of `/api/status`.

## Supported Meinberg LTOS Devices

The exporter has been tested with the following Meinberg LTOS devices:

| Model | Receiver | LTOS Version     |
| ----- | -------- | ------------     |
| M600  | grc180   | 7.10.008 & later |
| M300  | pzf511   | 7.06.014-light   |

## Docker

```sh
docker run --rm -p 10123:10123 \
  -e MEINBERG_LTOS_EXPORTER_TARGET=https://<device> \
  -e MEINBERG_LTOS_EXPORTER_AUTH_USER=<user> \
  -e MEINBERG_LTOS_EXPORTER_AUTH_PASS=<password> \
  ghcr.io/raphaelthomas/meinberg_ltos_exporter:latest
```

## Configuration

The exporter can be configured via the following parameters:

``` sh
usage: meinberg_ltos_exporter --target=TARGET [<flags>]

Prometheus exporter for Meinberg LTOS devices


Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --[no-]version             Show application version.
      --web.listen-address=":10123"
                                 Address to listen on for web interface and telemetry ($MEINBERG_LTOS_EXPORTER_LISTEN_ADDRESS)
      --web.telemetry-path="/metrics"
                                 Path under which to expose metrics ($MEINBERG_LTOS_EXPORTER_METRICS_PATH)
      --target=TARGET            Base URL of the Meinberg LTOS device (e.g. https://clock.example.com) ($MEINBERG_LTOS_EXPORTER_TARGET)
      --auth-user=AUTH-USER      Basic auth username (prefer env var over CLI flag) ($MEINBERG_LTOS_EXPORTER_AUTH_USER)
      --auth-pass=AUTH-PASS      Basic auth password (prefer env var over CLI flag) ($MEINBERG_LTOS_EXPORTER_AUTH_PASS)
      --timeout=5s               Timeout for HTTP requests to Meinberg device ($MEINBERG_LTOS_EXPORTER_TIMEOUT)
      --[no-]ignore-ssl-verify   Ignore SSL certificate verification ($MEINBERG_LTOS_EXPORTER_IGNORE_SSL_VERIFY)
      --log-level=info           Log level (debug, info, warn, error)
      --[no-]collector.system    Enable system collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_SYSTEM)
      --[no-]collector.notification
                                 Enable notification collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_NOTIFICATION)
      --[no-]collector.network   Enable network collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_NETWORK)
      --[no-]collector.storage   Enable storage collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_STORAGE)
      --[no-]collector.clock     Enable clock collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_CLOCK)
      --[no-]collector.receiver  Enable receiver collectors (GNSS + DCF77). ($MEINBERG_LTOS_EXPORTER_COLLECTOR_RECEIVER)
      --[no-]collector.ntp       Enable NTP collector. ($MEINBERG_LTOS_EXPORTER_COLLECTOR_NTP)
```

These parameters can be provided as environment variables or command-line
arguments.

### Authentication

The exporter supports Basic Authentication. Ensure the user has the "info"
access level (lowest permission level) configured on the LTOS device.

## Build

To build the exporter, run the following command, which will create an
executable named `meinberg_ltos_exporter`:

```sh
make build
```

or run `goreleaser` directly:

```sh
goreleaser build --snapshot --clean
```

## Local Development

### Adding new test data files

1. Save the API JSON response in `tests/testdata/` using a descriptive name (e.g. `m600-gps.json`).
1. Generate and validate the expected metrics output using the following:

   ```
   UPDATE_GOLDEN=1 make test
   ```

1. Open a PR with the new test data file and the updated golden file. CI will validate the new test data and metrics output.

### Running with Mock API

Run the following in three separate terminal windows:

1. A mock LTOS API serving static test data. Supply a file from
`tests/testdata/` as static API JSON response via `FILE`:

   ```sh
   make mock-api FILE=tests/testdata/m600-gps.json AUTH_USER=myuser AUTH_PASS=mypass
   ```

1. Run the exporter (triggers a build if necessary):

   ```sh
   make run AUTH_USER=myuser AUTH_PASS=mypass
   ```

1. Scrape the metrics exposed by the exporter:

   ```sh
   curl -s http://localhost:10123/metrics | grep meinberg_ltos
   ```
