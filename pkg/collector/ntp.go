package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const (
	ntpSysSubsystem  = "ntp_sys"
	ntpPeerSubsystem = "ntp_peer"
)

var (
	variableLabelsNTPSys   = []string{"host", "refid"}
	variableLabelsNTPPeers = []string{"host", "refid", "peer_name", "peer_address"}
)

var (
	ntpSysStratum = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "stratum"),
			"Meinberg NTP stratum level",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysPrecision = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "precision_seconds"),
			"Meinberg NTP precision in seconds",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysRootDelay = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "root_delay_seconds"),
			"Meinberg NTP root delay in seconds",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysRootDispersion = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "root_dispersion_seconds"),
			"Meinberg NTP root dispersion in seconds",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysClockJitter = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "clock_jitter_seconds"),
			"Meinberg NTP clock jitter in seconds",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysClockWander = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "clock_wander_seconds_per_second"),
			"Meinberg NTP clock wander in seconds per second",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysLeapIndicator = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "leap_indicator"),
			"Meinberg NTP leap indicator (0 = no warning, 1 = last minute has 61 seconds, 2 = last minute has 59 seconds, 3 = unknown)",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpSysLeapSecond = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpSysSubsystem, "leap_second_timestamp_seconds"),
			"Meinberg NTP leap second (last or next) in seconds since UNIX epoch",
			variableLabelsNTPSys,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpPeerOffset = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpPeerSubsystem, "offset_seconds"),
			"Meinberg NTP peer offset in seconds",
			variableLabelsNTPPeers,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpPeerDelay = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpPeerSubsystem, "delay_seconds"),
			"Meinberg NTP peer delay in seconds",
			variableLabelsNTPPeers,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpPeerDispersion = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpPeerSubsystem, "dispersion_seconds"),
			"Meinberg NTP peer dispersion in seconds",
			variableLabelsNTPPeers,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpPeerSynchronized = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpPeerSubsystem, "synchronized"),
			"Meinberg NTP peer synchronized state (1 if synchronized, 0 otherwise)",
			variableLabelsNTPPeers,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
	ntpPeerLeapIndicator = typedDesc{
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(MetricNamespace, ntpPeerSubsystem, "leap_indicator"),
			"Meinberg NTP peer leap indicator (0 = no warning, 1 = last minute has 61 seconds, 2 = last minute has 59 seconds, 3 = unknown)",
			variableLabelsNTPPeers,
			nil,
		),
		valueType: prometheus.GaugeValue,
	}
)

func describeNTP(ch chan<- *prometheus.Desc) {
	describeNTPSys(ch)
	describeNTPPeers(ch)
}

func describeNTPSys(ch chan<- *prometheus.Desc) {
	ch <- ntpSysStratum.desc
	ch <- ntpSysPrecision.desc
	ch <- ntpSysRootDelay.desc
	ch <- ntpSysRootDispersion.desc
	ch <- ntpSysClockJitter.desc
	ch <- ntpSysClockWander.desc
	ch <- ntpSysLeapIndicator.desc
	ch <- ntpSysLeapSecond.desc
}

func describeNTPPeers(ch chan<- *prometheus.Desc) {
	ch <- ntpPeerOffset.desc
	ch <- ntpPeerDelay.desc
	ch <- ntpPeerDispersion.desc
	ch <- ntpPeerLeapIndicator.desc
	ch <- ntpPeerSynchronized.desc
}

func (c *Collector) collectNTP(ch chan<- prometheus.Metric, host string, assocs []models.NTPAssociation) {
	for _, a := range assocs {
		if a.IsSys() {
			c.collectNTPSysAssoc(ch, host, a)
		} else {
			c.collectNTPPeerAssoc(ch, host, a)
		}
	}
}

func (c *Collector) collectNTPSysAssoc(ch chan<- prometheus.Metric, host string, assoc models.NTPAssociation) {
	labels := []string{host, assoc.RefID}

	ch <- ntpSysStratum.mustNewConstMetric(assoc.Stratum, labels...)
	ch <- ntpSysPrecision.mustNewConstMetric(assoc.PrecisionSeconds(), labels...)
	ch <- ntpSysRootDelay.mustNewConstMetric(assoc.RootDelay, labels...)
	ch <- ntpSysRootDispersion.mustNewConstMetric(assoc.RootDispersion, labels...)
	ch <- ntpSysClockJitter.mustNewConstMetric(assoc.ClockJitter, labels...)
	ch <- ntpSysClockWander.mustNewConstMetric(assoc.ClockWander, labels...)
	ch <- ntpSysLeapIndicator.mustNewConstMetric(float64(assoc.LeapIndicator), labels...)
	ch <- ntpSysLeapSecond.mustNewConstMetric(float64(assoc.LeapSecondUnix), labels...)
}

func (c *Collector) collectNTPPeerAssoc(ch chan<- prometheus.Metric, host string, assoc models.NTPAssociation) {
	labels := []string{host, assoc.RefID, assoc.Name, assoc.Address}
	if assoc.Offset != nil {
		ch <- ntpPeerOffset.mustNewConstMetric(*assoc.Offset, labels...)
	}
	if assoc.Delay != nil {
		ch <- ntpPeerDelay.mustNewConstMetric(*assoc.Delay, labels...)
	}
	if assoc.Dispersion != nil {
		ch <- ntpPeerDispersion.mustNewConstMetric(*assoc.Dispersion, labels...)
	}
	ch <- ntpPeerLeapIndicator.mustNewConstMetric(float64(assoc.LeapIndicator), labels...)
	ch <- ntpPeerSynchronized.mustNewConstMetric(boolToFloat64(assoc.LeapIndicator != models.Unknown), labels...)
}
