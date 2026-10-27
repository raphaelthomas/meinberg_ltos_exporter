package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const rcvGNSSSubsystem = "clock_receiver_gnss"

var (
	clkRcvGNSSSatInView = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "satellites_in_view"),
			"Number of satellites (theoretically) in view of the GNSS receiver",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSSatGood = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "satellites_good"),
			"Number of good satellites for the GNSS receiver",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSLatitude = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "latitude_degrees"),
			"Meinberg GNSS receiver latitude",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSLongitude = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "longitude_degrees"),
			"Meinberg GNSS receiver longitude",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSAltitude = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "altitude_meters"),
			"Meinberg GNSS receiver altitude",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSAntConnected = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "antenna_connected"),
			"Meinberg GNSS receiver antenna connected (1 = connected, 0 = not connected)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSAntShortCircuit = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "antenna_short_circuit"),
			"Meinberg GNSS receiver antenna short circuit detected (1 = short circuit, 0 = no short circuit)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSSynced = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "synchronized"),
			"Meinberg GNSS receiver synchronization status (1 = synced, 0 = not synced)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSTracking = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "tracking"),
			"Meinberg GNSS receiver tracking status (1 = tracking, 0 = not tracking)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSColdBoot = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "cold_boot"),
			"GNSS receiver cold boot status (1 = cold boot, 0 = not cold boot)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	clkRcvGNSSWarmBoot = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, rcvGNSSSubsystem, "warm_boot"),
			"GNSS receiver warm boot status (1 = warm boot, 0 = not warm boot)",
			[]string{"host", "clock_id"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeReceiverGNSS(ch chan<- *prometheus.Desc) {
	ch <- clkRcvGNSSSatInView.desc
	ch <- clkRcvGNSSSatGood.desc
	ch <- clkRcvGNSSLatitude.desc
	ch <- clkRcvGNSSLongitude.desc
	ch <- clkRcvGNSSAltitude.desc
	ch <- clkRcvGNSSAntConnected.desc
	ch <- clkRcvGNSSAntShortCircuit.desc
	ch <- clkRcvGNSSSynced.desc
	ch <- clkRcvGNSSTracking.desc
	ch <- clkRcvGNSSColdBoot.desc
	ch <- clkRcvGNSSWarmBoot.desc
}

func (c *Collector) collectReceiverGNSS(ch chan<- prometheus.Metric, host string, slots []models.Slot) {
	forEachClockSlot(slots, func(slot models.Slot) {
		if slot.Module.Satellites != nil {
			ch <- clkRcvGNSSSatInView.mustNewConstMetric(slot.Module.Satellites.InView, host, slot.Name)
			ch <- clkRcvGNSSSatGood.mustNewConstMetric(slot.Module.Satellites.Good, host, slot.Name)
			ch <- clkRcvGNSSLatitude.mustNewConstMetric(slot.Module.Satellites.Latitude, host, slot.Name)
			ch <- clkRcvGNSSLongitude.mustNewConstMetric(slot.Module.Satellites.Longitude, host, slot.Name)
			ch <- clkRcvGNSSAltitude.mustNewConstMetric(slot.Module.Satellites.Altitude, host, slot.Name)
		}

		if slot.Module.GRC != nil {
			if slot.Module.GRC.Antenna != nil {
				ch <- clkRcvGNSSAntConnected.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Antenna.IsConnected), host, slot.Name)
				ch <- clkRcvGNSSAntShortCircuit.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Antenna.HasShortCircuit), host, slot.Name)
			}

			if slot.Module.GRC.Receiver != nil {
				ch <- clkRcvGNSSSynced.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsSynchronized), host, slot.Name)
				ch <- clkRcvGNSSTracking.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsTracking), host, slot.Name)
				ch <- clkRcvGNSSWarmBoot.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsWarmBooting), host, slot.Name)
				ch <- clkRcvGNSSColdBoot.mustNewConstMetric(boolToFloat64(slot.Module.GRC.Receiver.IsColdBooting), host, slot.Name)
			}
		}
	})
}
