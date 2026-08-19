package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

const (
	firmwareVersionPrefix = "fw_"
	firmwareImagePrefix   = "firmware-"
)

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (td typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(td.desc, td.valueType, value, labels...)
}

// firmwareVersion strips the prefix of raw form "fw_7.06.014-light". The
// edition suffix is kept, as it distinguishes firmware builds.
func firmwareVersion(version string) string {
	return strings.TrimPrefix(version, firmwareVersionPrefix)
}

// firmwareArch derives the architecture from the image of raw form
// "firmware-7.06.014-light-x86". The running version anchors the match, as
// the version itself may contain the separator.
func firmwareArch(image, running string) string {
	arch, ok := strings.CutPrefix(image, firmwareImagePrefix+firmwareVersion(running)+"-")
	if !ok {
		return ""
	}
	return arch
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func forEachSlotWithModule(slots []models.Slot, slotType string, fn func(models.Slot)) {
	for _, slot := range slots {
		if slot.Type != slotType || slot.Module == nil {
			continue
		}
		fn(slot)
	}
}

func forEachCPUSlot(slots []models.Slot, fn func(models.Slot)) {
	forEachSlotWithModule(slots, models.SlotTypeCPU, fn)
}

func forEachClockSlot(slots []models.Slot, fn func(models.Slot)) {
	forEachSlotWithModule(slots, models.SlotTypeClock, fn)
}
