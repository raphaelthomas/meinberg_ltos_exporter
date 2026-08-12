// Copyright 2026 Raphael Seebacher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/buildinfo"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/collector"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi"
)

// Config holds the exporter configuration
type Config struct {
	ListenAddress   string
	MetricsPath     string
	Target          string
	LogLevel        slog.Level
	AuthBasicUser   string
	AuthBasicPass   string
	IgnoreSSLVerify bool
	Collector       collector.Config
}

// parseFlags parses command-line flags using kingpin
func parseFlags() *Config {
	app := kingpin.New("meinberg_ltos_exporter", "Prometheus exporter for Meinberg LTOS devices")
	app.Version(buildinfo.Version)
	app.HelpFlag.Short('h')

	cfg := &Config{}

	const envPrefix = "MEINBERG_LTOS_EXPORTER_"

	app.Flag("web.listen-address", "Address to listen on for web interface and telemetry").
		Default(":10123").
		Envar(envPrefix + "LISTEN_ADDRESS").
		StringVar(&cfg.ListenAddress)

	app.Flag("web.telemetry-path", "Path under which to expose metrics").
		Default("/metrics").
		Envar(envPrefix + "METRICS_PATH").
		StringVar(&cfg.MetricsPath)

	app.Flag("target", "Base URL of the Meinberg LTOS device (e.g. https://clock.example.com)").
		Required().
		Envar(envPrefix + "TARGET").
		StringVar(&cfg.Target)

	app.Flag("auth-user", "Basic auth username (prefer env var over CLI flag)").
		Envar(envPrefix + "AUTH_USER").
		StringVar(&cfg.AuthBasicUser)

	app.Flag("auth-pass", "Basic auth password (prefer env var over CLI flag)").
		Envar(envPrefix + "AUTH_PASS").
		StringVar(&cfg.AuthBasicPass)

	app.Flag("timeout", "Timeout for HTTP requests to Meinberg device").
		Default("5s").
		Envar(envPrefix + "TIMEOUT").
		DurationVar(&cfg.Collector.Timeout)

	app.Flag("ignore-ssl-verify", "Ignore SSL certificate verification").
		Default("false").
		Envar(envPrefix + "IGNORE_SSL_VERIFY").
		BoolVar(&cfg.IgnoreSSLVerify)

	logLevelFlag := app.Flag("log-level", "Log level (debug, info, warn, error)").
		Default("info").
		Enum("debug", "info", "warn", "error")

	app.Flag("collector.system", "Enable system collector.").
		Default("true").
		Envar(envPrefix + "COLLECTOR_SYSTEM").
		BoolVar(&cfg.Collector.System)

	app.Flag("collector.notification", "Enable notification collector.").
		Default("true").
		Envar(envPrefix + "COLLECTOR_NOTIFICATION").
		BoolVar(&cfg.Collector.Notification)

	app.Flag("collector.network", "Enable network collector.").
		Default("true").
		Envar(envPrefix + "COLLECTOR_NETWORK").
		BoolVar(&cfg.Collector.Network)

	app.Flag("collector.storage", "Enable storage collector.").
		Default("true").
		Envar(envPrefix + "COLLECTOR_STORAGE").
		BoolVar(&cfg.Collector.Storage)

	app.Flag("collector.clock", "Enable clock collector.").
		Default("true").
		Envar(envPrefix + "COLLECTOR_CLOCK").
		BoolVar(&cfg.Collector.Clock)

	app.Flag("collector.receiver", "Enable receiver collectors (GNSS + DCF77).").
		Default("true").
		Envar(envPrefix + "COLLECTOR_RECEIVER").
		BoolVar(&cfg.Collector.Receiver)

	app.Flag("collector.ntp", "Enable NTP collector.").
		Default("true").
		Envar(envPrefix + "COLLECTOR_NTP").
		BoolVar(&cfg.Collector.NTP)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	if err := cfg.LogLevel.UnmarshalText([]byte(*logLevelFlag)); err != nil {
		cfg.LogLevel = slog.LevelInfo
	}

	return cfg
}

func main() {
	cfg := parseFlags()

	logLevel := &slog.LevelVar{}
	logLevel.Set(cfg.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	client, err := ltosapi.NewClient(cfg.Target, cfg.AuthBasicUser, cfg.AuthBasicPass, cfg.IgnoreSSLVerify)
	if err != nil {
		logger.Error("failed to create LTOS API client", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting Meinberg LTOS Exporter",
		"version", buildinfo.Version,
		"listen_address", cfg.ListenAddress,
		"target", client.Target(),
	)

	if cfg.IgnoreSSLVerify {
		logger.Warn("TLS certificate verification disabled via --ignore-ssl-verify")
	}

	prometheus.MustRegister(collector.NewCollector(cfg.Collector, client, logger))
	prometheus.MustRegister(versioncollector.NewCollector(prometheus.BuildFQName(collector.MetricNamespace, "", "exporter")))

	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.Handler())

	landingPageData := struct {
		Target      string
		MetricsPath string
	}{
		Target:      cfg.Target,
		MetricsPath: cfg.MetricsPath,
	}

	landingPageTmpl := template.Must(template.New("landingPage").Parse(`
<!DOCTYPE html>
<html>
<head>
  <title>Meinberg LTOS Exporter</title>
</head>
<body>
  <h1>Meinberg LTOS Exporter</h1>
  <p>Prometheus exporter for Meinberg LTOS devices.</p>
	<p>Check <a href="{{.MetricsPath}}">{{.MetricsPath}}</a> for the Prometheus metrics in text exposition format scraped from {{.Target}}.</p>
</body>
</html>
`))

	if cfg.MetricsPath != "/" && cfg.MetricsPath != "" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			if err := landingPageTmpl.Execute(w, landingPageData); err != nil {
				logger.Error("Failed to write response", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		})
	}

	srv := &http.Server{Addr: cfg.ListenAddress, Handler: mux}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		logger.Info("Received signal, shutting down HTTP server", "signal", sig)
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("HTTP server shutdown error", "error", err)
		}
	}()

	logger.Info("HTTP server listening", "address", cfg.ListenAddress)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
