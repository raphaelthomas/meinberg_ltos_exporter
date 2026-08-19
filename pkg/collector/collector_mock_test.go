package collector_test

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/collector"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

type mockFetcher struct {
	target   string
	response *models.StatusResponse
	err      error
}

func (m *mockFetcher) FetchStatus(_ context.Context, _ *slog.Logger) (*models.StatusResponse, error) {
	return m.response, m.err
}

func (m *mockFetcher) Target() string {
	return m.target
}

func allEnabledConfig() collector.Config {
	return collector.Config{
		Timeout:      5 * time.Second,
		System:       true,
		Notification: true,
		Network:      true,
		Storage:      true,
		Clock:        true,
		Receiver:     true,
		NTP:          true,
	}
}

// collectMetrics gathers all metrics from the collector and returns them
// as a map of metric name to list of sample lines.
func collectMetrics(t *testing.T, c *collector.Collector) map[string][]string {
	t.Helper()
	raw := gatherMetrics(t, c)
	result := make(map[string][]string)
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		name := metricName(line)
		result[name] = append(result[name], line)
	}
	return result
}

func TestCollect_DeviceUnreachable(t *testing.T) {
	mock := &mockFetcher{
		target: "https://unreachable.example.com",
		err:    fmt.Errorf("connection refused"),
	}
	c := collector.NewCollector(allEnabledConfig(), mock, slog.New(slog.DiscardHandler))
	metrics := collectMetrics(t, c)

	// Must report up=0
	upSamples := metrics["meinberg_ltos_up"]
	if len(upSamples) != 1 {
		t.Fatalf("expected 1 up sample, got %d", len(upSamples))
	}
	if !strings.HasSuffix(upSamples[0], " 0") {
		t.Errorf("expected up=0, got: %s", upSamples[0])
	}

	// Must still report scrape duration
	if _, ok := metrics["meinberg_ltos_scrape_duration_seconds"]; !ok {
		t.Error("expected scrape_duration_seconds metric even on failure")
	}

	// Must NOT report build_info (device was unreachable)
	if _, ok := metrics["meinberg_ltos_build_info"]; ok {
		t.Error("did not expect build_info when device is unreachable")
	}
}

func TestCollect_DeviceReachable(t *testing.T) {
	mock := &mockFetcher{
		target: "https://clock.example.com",
		response: &models.StatusResponse{
			SystemInformation: models.SystemInformation{
				Hostname: "clock1",
				Version:  "fw_7.10.008",
				Model:    "M600",
			},
			Data: models.StatusData{
				RestAPI: models.RestAPI{Version: "20.05.013"},
				System: models.System{
					UptimeSeconds: 12345,
					CPULoad:       models.CPULoad{Load1: 0.1, Load5: 0.2, Load15: 0.3},
					Memory:        models.Memory{Total: 228428 * 1024, Free: 161732 * 1024},
					Firmware: models.Firmware{
						Running: "fw_7.10.008",
						Image:   "firmware-7.10.008-x32",
					},
				},
			},
		},
	}

	c := collector.NewCollector(allEnabledConfig(), mock, slog.New(slog.DiscardHandler))
	metrics := collectMetrics(t, c)

	// Must report up=1
	upSamples := metrics["meinberg_ltos_up"]
	if len(upSamples) != 1 {
		t.Fatalf("expected 1 up sample, got %d", len(upSamples))
	}
	if !strings.HasSuffix(upSamples[0], " 1") {
		t.Errorf("expected up=1, got: %s", upSamples[0])
	}

	// Must report build_info, with the prefix stripped and the arch derived
	buildInfo := metrics["meinberg_ltos_build_info"]
	if len(buildInfo) != 1 {
		t.Fatalf("expected 1 build_info sample, got %d", len(buildInfo))
	}
	for _, want := range []string{`firmware_version="7.10.008"`, `firmware_arch="x32"`} {
		if !strings.Contains(buildInfo[0], want) {
			t.Errorf("expected build_info to contain %s, got: %s", want, buildInfo[0])
		}
	}

	// Must report system metrics
	if _, ok := metrics["meinberg_ltos_system_uptime_seconds"]; !ok {
		t.Error("expected system_uptime_seconds metric")
	}
}

func TestCollect_CollectorsDisabled(t *testing.T) {
	mock := &mockFetcher{
		target: "https://clock.example.com",
		response: &models.StatusResponse{
			SystemInformation: models.SystemInformation{Hostname: "clock1"},
			Data: models.StatusData{
				RestAPI: models.RestAPI{Version: "20.05.013"},
				System: models.System{
					UptimeSeconds: 100,
					CPULoad:       models.CPULoad{Load1: 0.1, Load5: 0.2, Load15: 0.3},
					Memory:        models.Memory{Total: 1024, Free: 512},
				},
				Notification: models.Notification{
					Events: []models.Event{{Type: "test", Name: "test-event"}},
				},
			},
		},
	}

	// Disable all subsystem collectors
	cfg := collector.Config{
		Timeout:      5 * time.Second,
		System:       false,
		Notification: false,
		Network:      false,
		Storage:      false,
		Clock:        false,
		Receiver:     false,
		NTP:          false,
	}

	c := collector.NewCollector(cfg, mock, slog.New(slog.DiscardHandler))
	metrics := collectMetrics(t, c)

	// Core metrics must still be present
	if _, ok := metrics["meinberg_ltos_up"]; !ok {
		t.Error("expected up metric even with all collectors disabled")
	}
	if _, ok := metrics["meinberg_ltos_build_info"]; !ok {
		t.Error("expected build_info even with all collectors disabled")
	}

	// Subsystem metrics must be absent
	subsystemPrefixes := []string{
		"meinberg_ltos_system_",
		"meinberg_ltos_notification_",
		"meinberg_ltos_network_port_",
		"meinberg_ltos_storage_",
		"meinberg_ltos_clock_",
		"meinberg_ltos_clock_receiver_gnss_",
		"meinberg_ltos_clock_receiver_dcf77_",
		"meinberg_ltos_ntp_sys_",
		"meinberg_ltos_ntp_peer_",
	}

	for name := range metrics {
		for _, prefix := range subsystemPrefixes {
			if strings.HasPrefix(name, prefix) {
				t.Errorf("unexpected subsystem metric %q when all collectors disabled", name)
			}
		}
	}
}
