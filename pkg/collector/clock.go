package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

const clockSubsystem = "clock"

var (
	clkInfo = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, clockSubsystem, "info"),
			"Meinberg clock module information as labels (model, serial number, software revision, oscillator type)",
			[]string{"host", "clock_id", "model", "serial_number", "software_revision", "oscillator_type"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkSyncStatus = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, clockSubsystem, "synchronized"),
			"Meinberg clock synchronization status (1 = synchronized, 0 = not synchronized)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkOscillatorWarmedUp = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, clockSubsystem, "oscillator_warmed_up"),
			"Meinberg clock oscillator warmed up status (1 = warmed up, 0 = not warmed up)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkEstTimeQuality = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, clockSubsystem, "estimated_time_quality_seconds"),
			"Estimated upper bound in seconds on the time quality of the clock",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeClock(ch chan<- *prometheus.Desc) {
	ch <- clkInfo.desc
	ch <- clkSyncStatus.desc
	ch <- clkOscillatorWarmedUp.desc
	ch <- clkEstTimeQuality.desc
}

func (c *Collector) collectClock(ch chan<- prometheus.Metric, host string, slots []models.Slot) {
	forEachClockSlot(slots, func(slot models.Slot) {
		oscillatorType := "unknown"
		if slot.Module.SyncStatus != nil {
			oscillatorType = slot.Module.SyncStatus.OscillatorType
			ch <- clkSyncStatus.mustNewConstMetric(boolToFloat64(slot.Module.SyncStatus.ClockStatus.IsSynchronized()), host, slot.Name)
			ch <- clkOscillatorWarmedUp.mustNewConstMetric(boolToFloat64(slot.Module.SyncStatus.ClockStatus.IsOscillatorWarmedUp()), host, slot.Name)
			if slot.Module.SyncStatus.TimeQuality != nil {
				ch <- clkEstTimeQuality.mustNewConstMetric(slot.Module.SyncStatus.TimeQuality.Seconds(), host, slot.Name)
			}
		}
		if slot.Module.Info != nil {
			ch <- clkInfo.mustNewConstMetric(1.0, host, slot.Name, slot.Module.Info.Model, slot.Module.Info.SerialNumber.String(), slot.Module.Info.SoftwareRevision, oscillatorType)
		}
	})
}
