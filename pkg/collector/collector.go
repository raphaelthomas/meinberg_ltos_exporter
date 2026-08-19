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

// Package collector implements the Prometheus collector for Meinberg LTOS metrics
package collector

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

const (
	MetricNamespace = "meinberg_ltos"
	rootSubsystem   = ""
)

var scrapeID atomic.Uint64

type Config struct {
	Timeout      time.Duration
	System       bool
	Notification bool
	Network      bool
	Storage      bool
	Clock        bool
	Receiver     bool
	NTP          bool
}

type StatusFetcher interface {
	FetchStatus(ctx context.Context, logger *slog.Logger) (*models.StatusResponse, error)
	Target() string
}

type Collector struct {
	config Config
	client StatusFetcher
	logger *slog.Logger

	up             typedDesc
	scrapeDuration typedDesc
	buildInfo      typedDesc
}

func NewCollector(config Config, client StatusFetcher, logger *slog.Logger) *Collector {
	if !config.System {
		logger.Info("Collector disabled", "collector", "system")
	}
	if !config.Notification {
		logger.Info("Collector disabled", "collector", "notification")
	}
	if !config.Network {
		logger.Info("Collector disabled", "collector", "network")
	}
	if !config.Storage {
		logger.Info("Collector disabled", "collector", "storage")
	}
	if !config.Clock {
		logger.Info("Collector disabled", "collector", "clock")
	}
	if !config.Receiver {
		logger.Info("Collector disabled", "collector", "receiver")
	}
	if !config.NTP {
		logger.Info("Collector disabled", "collector", "ntp")
	}

	return &Collector{
		config: config,
		client: client,
		logger: logger,
		up: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, rootSubsystem, "up"),
				"Indicates if the Meinberg LTOS device is reachable (1 = up, 0 = down)",
				[]string{"target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		scrapeDuration: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, rootSubsystem, "scrape_duration_seconds"),
				"Duration of the scrape of the Meinberg LTOS device in seconds",
				[]string{"target"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
		buildInfo: typedDesc{
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(MetricNamespace, rootSubsystem, "build_info"),
				"Meinberg device build information as labels (e.g., API version, firmware version, firmware architecture, host)",
				[]string{"target", "host", "api_version", "firmware_version", "firmware_arch"},
				nil,
			),
			valueType: prometheus.GaugeValue,
		},
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up.desc
	ch <- c.scrapeDuration.desc
	ch <- c.buildInfo.desc

	if c.config.System {
		describeSystem(ch)
	}
	if c.config.Notification {
		describeNotification(ch)
	}
	if c.config.Network {
		describeNetwork(ch)
	}
	if c.config.Storage {
		describeStorage(ch)
	}
	if c.config.Clock {
		describeClock(ch)
	}
	if c.config.Receiver {
		describeReceiverGNSS(ch)
		describeReceiverDCF77(ch)
	}
	if c.config.NTP {
		describeNTP(ch)
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	logger := c.logger.With("scrape_id", scrapeID.Add(1))

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	start := time.Now()
	up := 0.0

	defer func() {
		seconds := time.Since(start).Seconds()
		ch <- c.scrapeDuration.mustNewConstMetric(seconds, c.client.Target())
		ch <- c.up.mustNewConstMetric(up, c.client.Target())
	}()

	logger.Debug("Collecting metrics from Meinberg LTOS device", "target", c.client.Target())

	status, err := c.client.FetchStatus(ctx, logger)
	if err != nil {
		logger.Warn("Failed to fetch Meinberg LTOS device status", "error", err)
		return
	}

	up = 1.0
	host := status.SystemInformation.Hostname

	firmware := status.Data.System.Firmware
	arch := firmwareArch(firmware.Image, firmware.Running)
	if arch == "" && firmware.Image != "" {
		logger.Debug("Failed to derive firmware architecture from firmware image",
			"firmware_image", firmware.Image, "running_firmware", firmware.Running)
	}
	ch <- c.buildInfo.mustNewConstMetric(1.0, c.client.Target(), host,
		status.Data.RestAPI.Version, firmwareVersion(status.SystemInformation.Version), arch)

	if c.config.System {
		c.collectSystem(ch, host, status.SystemInformation, status.Data.System, status.Data.Chassis.Slots)
	}
	if c.config.Notification {
		c.collectNotification(ch, host, status.Data.Notification.Events)
	}
	if c.config.Network {
		c.collectNetwork(ch, host, status.Data.Network)
	}
	if c.config.Storage {
		c.collectStorage(ch, host, status.Data.System.Mounts)
	}
	if c.config.NTP {
		c.collectNTP(ch, host, status.Data.NTP)
	}
	if c.config.Clock {
		c.collectClock(ch, host, status.Data.Chassis.Slots)
	}
	if c.config.Receiver {
		c.collectReceiverGNSS(ch, host, status.Data.Chassis.Slots)
		c.collectReceiverDCF77(ch, host, status.Data.Chassis.Slots)
	}

	logger.Debug("Done collecting metrics from Meinberg LTOS device", "target", c.client.Target(), "host", host)
}
