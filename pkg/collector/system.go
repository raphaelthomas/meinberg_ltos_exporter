package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const systemSubsystem = "system"

var (
	systemInfo = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, systemSubsystem, "info"),
			"Meinberg system information as labels (e.g., model, serial number, host)",
			[]string{"host", "model", "serial_number"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemCPUInfo = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, systemSubsystem, "cpu_info"),
			"CPU information as labels (model, serial, etc.)",
			[]string{"host", "model", "serial_number"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemUptimeSeconds = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, systemSubsystem, "uptime_seconds"),
			"System uptime in seconds",
			[]string{"host"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemCPULoadAvg = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, systemSubsystem, "cpu_load_avg"),
			"CPU load averaged over 1, 5, and 15 minutes",
			[]string{"host", "period"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemMemoryBytes = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, systemSubsystem, "memory_bytes"),
			"Total memory in bytes",
			[]string{"host"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	systemMemoryFreeBytes = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, systemSubsystem, "memory_free_bytes"),
			"Free memory in bytes",
			[]string{"host"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeSystem(ch chan<- *prometheus.Desc) {
	ch <- systemInfo.desc
	ch <- systemCPUInfo.desc
	ch <- systemUptimeSeconds.desc
	ch <- systemCPULoadAvg.desc
	ch <- systemMemoryBytes.desc
	ch <- systemMemoryFreeBytes.desc
}

func (c *Collector) collectSystem(ch chan<- prometheus.Metric, host string, systemInformation models.SystemInformation, system models.System, slots []models.Slot) {
	ch <- systemInfo.mustNewConstMetric(1.0, host, systemInformation.Model, systemInformation.SerialNumber.String())
	ch <- systemUptimeSeconds.mustNewConstMetric(system.UptimeSeconds, host)
	ch <- systemCPULoadAvg.mustNewConstMetric(system.CPULoad.Load1, host, "1")
	ch <- systemCPULoadAvg.mustNewConstMetric(system.CPULoad.Load5, host, "5")
	ch <- systemCPULoadAvg.mustNewConstMetric(system.CPULoad.Load15, host, "15")
	ch <- systemMemoryBytes.mustNewConstMetric(system.Memory.Total, host)
	ch <- systemMemoryFreeBytes.mustNewConstMetric(system.Memory.Free, host)

	forEachCPUSlot(slots, func(slot models.Slot) {
		if slot.Module.Info == nil {
			return
		}
		ch <- systemCPUInfo.mustNewConstMetric(1.0, host, slot.Module.Info.Model, slot.Module.Info.SerialNumber.String())
	})
}
