package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const storageSubsystem = "storage"

var (
	storageTotal = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, storageSubsystem, "total_bytes"),
			"Total size of the storage volume in bytes",
			[]string{"host", "mount"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	storageUsed = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, storageSubsystem, "used_bytes"),
			"Used bytes of the storage volume",
			[]string{"host", "mount"},
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeStorage(ch chan<- *prometheus.Desc) {
	ch <- storageTotal.desc
	ch <- storageUsed.desc
}

func (c *Collector) collectStorage(ch chan<- prometheus.Metric, host string, mounts []models.Mount) {
	for _, mount := range mounts {
		ch <- storageTotal.mustNewConstMetric(mount.Size, host, mount.Mountpoint)
		ch <- storageUsed.mustNewConstMetric(mount.Used, host, mount.Mountpoint)
	}
}
