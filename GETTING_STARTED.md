# Getting Started with the `meinberg_ltos_exporter`

This guide gets you from zero to scraping Meinberg LTOS metrics in minutes.

## Prerequisites

- A Meinberg device running LANTIME OS reachable over the network.
- A user on the device with **info** access level (== lowest permission level).
- Docker **or** a Linux/macOS/Windows machine with Internet access (needed to
download the exporter)

## Step 0: Setup environment variables

Make the necessary minimal configuration available as environment variables:

```sh
export MEINBERG_LTOS_EXPORTER_TARGET="https://192.0.2.123"
export MEINBERG_LTOS_EXPORTER_AUTH_USER="INFO_LEVEL_API_USER"
printf "Password: " && read -rs MEINBERG_LTOS_EXPORTER_AUTH_PASS && export MEINBERG_LTOS_EXPORTER_AUTH_PASS
```

> [!NOTE]
> `read -rs` reads the password from `stdin` without echoing it, then exports
> it into the current shell session. This is to not leak the password, e.g. to
> shell history.

> [!IMPORTANT]
> All of the following commands must be run from the same shell (or make sure
> that the above environment variables are available in each shell).

### Self-signed SSL certificates

If the Web interface of your Meinberg device uses a self-signed certificate,
make sure to additionally include the following environment variable:

```sh
export MEINBERG_LTOS_EXPORTER_IGNORE_SSL_VERIFY="true"
```

Run the `curl` command below with the `--insecure` flag to skip SSL
verification too.

## Step 1: Verify device connectivity

Run the following command to verify that the Meinberg device is accessible with
the configured user and password. Uncomment the `--insecure` flag if your
device uses a self-signed certificate.

```sh
# Add --insecure if your device uses a self-signed certificate.
curl \
  --user "$MEINBERG_LTOS_EXPORTER_AUTH_USER:$MEINBERG_LTOS_EXPORTER_AUTH_PASS" \
  "$MEINBERG_LTOS_EXPORTER_TARGET/api/status"
```

If you see a JSON response in the terminal, proceed to step 2.

## Step 2: Run the exporter

### Option A: Docker

Start the container, using the previously exported environment variables:

```sh
docker run --rm -p 10123:10123 \
  -e MEINBERG_LTOS_EXPORTER_TARGET \
  -e MEINBERG_LTOS_EXPORTER_AUTH_USER \
  -e MEINBERG_LTOS_EXPORTER_AUTH_PASS \
  -e MEINBERG_LTOS_EXPORTER_IGNORE_SSL_VERIFY \
  ghcr.io/raphaelthomas/meinberg_ltos_exporter:latest
```

### Option B: Binary

Download the latest release for your platform from the [releases
page](https://github.com/raphaelthomas/meinberg-ltos-exporter/releases/latest),
extract the archive, and then run:

```sh
./meinberg_ltos_exporter
```

## Step 3: Verify metrics are exposed by the exporter

```sh
curl -s http://localhost:10123/metrics | grep meinberg_ltos
```

You should see lines like:

```
# HELP meinberg_ltos_build_info Meinberg device build information as labels (e.g., API version, firmware version, firmware architecture, host)
# TYPE meinberg_ltos_build_info gauge
meinberg_ltos_build_info{api_version="20.05.013",firmware_arch="x32",firmware_version="7.10.008",host="mbg1.time.example.com",target="https://mbg1.time.example.com"} 1

[...]

# HELP meinberg_ltos_up Indicates if the Meinberg LTOS device is reachable (1 = up, 0 = down)
# TYPE meinberg_ltos_up gauge
meinberg_ltos_up{target="https://mbg1.time.example.com"} 1
```

## Step 4: Have Prometheus scrape the exporter

Add the following scrape job to your `prometheus.yml`, assuming Prometheus runs
on the same machine:

```yaml
scrape_configs:
  - job_name: meinberg_ltos
    static_configs:
      - targets: ["localhost:10123"]
```
