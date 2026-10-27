package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const rcvDCF77Subsystem = "clock_receiver_dcf77"

var (
	clkRcvDCF77FieldStrength = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvDCF77Subsystem, "field_strength"),
			"DCF77 receiver field strength",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvDCF77Correlation = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvDCF77Subsystem, "correlation"),
			"DCF77 receiver correlation",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeReceiverDCF77(ch chan<- *prometheus.Desc) {
	ch <- clkRcvDCF77FieldStrength.desc
	ch <- clkRcvDCF77Correlation.desc
}

func (c *Collector) collectReceiverDCF77(ch chan<- prometheus.Metric, host string, slots []models.Slot) {
	forEachClockSlot(slots, func(slot models.Slot) {
		if slot.Module.DCF77 == nil {
			return
		}
		ch <- clkRcvDCF77FieldStrength.mustNewConstMetric(slot.Module.DCF77.FieldStrength, host, slot.Name)
		ch <- clkRcvDCF77Correlation.mustNewConstMetric(slot.Module.DCF77.Correlation, host, slot.Name)
	})
}
