package collector

import (
	"testing"

	"github.com/raphaelthomas/meinberg-ltos-exporter/pkg/ltosapi/models"
)

func TestForEachSlotWithModule(t *testing.T) {
	slots := []models.Slot{
		{Type: models.SlotTypeCPU, Name: "cpu0", Module: &models.SlotModule{}},
		{Type: models.SlotTypeClock, Name: "clk0", Module: &models.SlotModule{}},
		{Type: models.SlotTypeCPU, Name: "cpu1", Module: &models.SlotModule{}},
		{Type: models.SlotTypeClock, Name: "clk1", Module: &models.SlotModule{}},
		{Type: models.SlotTypeCPU, Name: "cpu2", Module: nil},
		{Type: models.SlotTypeClock, Name: "clk2", Module: nil},
	}

	tests := []struct {
		name     string
		slots    []models.Slot
		fn       func([]models.Slot, func(models.Slot))
		expected []string
	}{
		{"multiple cpu slots", slots, forEachCPUSlot, []string{"cpu0", "cpu1"}},
		{"multiple clock slots", slots, forEachClockSlot, []string{"clk0", "clk1"}},
		{"empty input", []models.Slot{}, forEachCPUSlot, nil},
		{"all nil modules", []models.Slot{
			{Type: models.SlotTypeClock, Name: "clk0", Module: nil},
			{Type: models.SlotTypeClock, Name: "clk1", Module: nil},
		}, forEachClockSlot, nil},
		{"no matching type", []models.Slot{
			{Type: models.SlotTypeCPU, Name: "cpu0", Module: &models.SlotModule{}},
		}, forEachClockSlot, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			tt.fn(tt.slots, func(s models.Slot) {
				got = append(got, s.Name)
			})
			if len(got) != len(tt.expected) {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("index %d: got %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestFirmwareVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"prefixed", "fw_7.10.008", "7.10.008"},
		{"prefixed with edition", "fw_7.06.014-light", "7.06.014-light"},
		{"without prefix", "7.10.008", "7.10.008"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firmwareVersion(tt.version); got != tt.expected {
				t.Errorf("firmwareVersion(%q) = %q, want %q", tt.version, got, tt.expected)
			}
		})
	}
}

func TestFirmwareArch(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		running  string
		expected string
	}{
		{"plain version", "firmware-7.10.008-x32", "fw_7.10.008", "x32"},
		{"version with edition", "firmware-7.06.014-light-x86", "fw_7.06.014-light", "x86"},
		{"running without prefix", "firmware-7.10.008-x32", "7.10.008", "x32"},
		{"image of another version", "firmware-7.10.007-x32", "fw_7.10.008", ""},
		{"image without arch", "firmware-7.10.008", "fw_7.10.008", ""},
		{"unexpected image format", "7.10.008-x32", "fw_7.10.008", ""},
		{"empty image", "", "fw_7.10.008", ""},
		{"empty running", "firmware-7.10.008-x32", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firmwareArch(tt.image, tt.running); got != tt.expected {
				t.Errorf("firmwareArch(%q, %q) = %q, want %q", tt.image, tt.running, got, tt.expected)
			}
		})
	}
}

func TestBoolToFloat64(t *testing.T) {
	if got := boolToFloat64(true); got != 1.0 {
		t.Errorf("boolToFloat64(true) = %v, want 1.0", got)
	}
	if got := boolToFloat64(false); got != 0.0 {
		t.Errorf("boolToFloat64(false) = %v, want 0.0", got)
	}
}
