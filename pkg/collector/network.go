package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const networkSubsystem = "network_port"

var (
	networkPortUp = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "up"),
			"Network port link status (1 = up, 0 = down)",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	networkPortInfo = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "info"),
			"Network port information as labels",
			[]string{"host", "port", "speed", "duplex", "mac_address", "card_name"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	networkPortRxBytes = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "rx_bytes_total"),
			"Total bytes received on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortTxBytes = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "tx_bytes_total"),
			"Total bytes transmitted on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortRxPackets = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "rx_packets_total"),
			"Total packets received on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortTxPackets = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "tx_packets_total"),
			"Total packets transmitted on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortRxErrors = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "rx_errors_total"),
			"Total receive errors on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortTxErrors = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "tx_errors_total"),
			"Total transmit errors on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortRxDropped = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "rx_dropped_total"),
			"Total received packets dropped on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
	networkPortTxDropped = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, networkSubsystem, "tx_dropped_total"),
			"Total transmitted packets dropped on the network port",
			[]string{"host", "port"},
			nil,
		),
		valueType: prometheus.CounterValue,
	}
)

func describeNetwork(ch chan<- *prometheus.Desc) {
	ch <- networkPortUp.desc
	ch <- networkPortInfo.desc
	ch <- networkPortRxBytes.desc
	ch <- networkPortTxBytes.desc
	ch <- networkPortRxPackets.desc
	ch <- networkPortTxPackets.desc
	ch <- networkPortRxErrors.desc
	ch <- networkPortTxErrors.desc
	ch <- networkPortRxDropped.desc
	ch <- networkPortTxDropped.desc
}

func (c *Collector) collectNetwork(ch chan<- prometheus.Metric, host string, network models.Network) {
	for _, port := range network.Ports {
		ch <- networkPortUp.mustNewConstMetric(boolToFloat64(port.Link), host, port.Name)

		// Port information and statistics are only relevant if the port is up, so we skip them if the link is down
		if !port.Link {
			continue
		}

		ch <- networkPortInfo.mustNewConstMetric(1.0, host, port.Name, port.Speed, port.Duplex, port.MACAddress, port.CardName)

		// Older API versions do not expose network port statistics
		if port.Statistics == nil {
			continue
		}

		s := port.Statistics
		ch <- networkPortRxBytes.mustNewConstMetric(s.RxBytes, host, port.Name)
		ch <- networkPortTxBytes.mustNewConstMetric(s.TxBytes, host, port.Name)
		ch <- networkPortRxPackets.mustNewConstMetric(s.RxPackets, host, port.Name)
		ch <- networkPortTxPackets.mustNewConstMetric(s.TxPackets, host, port.Name)
		ch <- networkPortRxErrors.mustNewConstMetric(s.RxErrors, host, port.Name)
		ch <- networkPortTxErrors.mustNewConstMetric(s.TxErrors, host, port.Name)
		ch <- networkPortRxDropped.mustNewConstMetric(s.RxDropped, host, port.Name)
		ch <- networkPortTxDropped.mustNewConstMetric(s.TxDropped, host, port.Name)
	}
}
