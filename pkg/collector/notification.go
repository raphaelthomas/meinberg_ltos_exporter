package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const notificationSubsystem = "notification"

var eventLastTriggered = typedDesc{
	desc: prometheus.NewDesc(
		prometheus.BuildFQName(MetricNamespace, notificationSubsystem, "last_triggered_seconds"),
		"When an event last occurred as seconds since UNIX epoch (0 if never triggered)",
		[]string{"host", "type", "event"},
		nil,
	),
	valueType: prometheus.GaugeValue,
}

func describeNotification(ch chan<- *prometheus.Desc) {
	ch <- eventLastTriggered.desc
}

func (c *Collector) collectNotification(ch chan<- prometheus.Metric, host string, events []models.Event) {
	for _, event := range events {
		ch <- eventLastTriggered.mustNewConstMetric(event.LastTriggeredUnix, host, event.Type, event.Name)
	}
}
