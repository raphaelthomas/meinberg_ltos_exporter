package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Chassis struct {
	BackplaneRevision string `json:"backplane-revision"`
	Slots             []Slot `json:"slots"`
}

type Slot struct {
	Type   string      `json:"slot-type"`
	Name   string      `json:"slot-id"`
	Module *SlotModule `json:"module,omitempty"`
}

const (
	SlotTypeCPU   = "cpu"
	SlotTypeClock = "clk"
)

type SlotModule struct {
	Info       *SlotModuleInfo `json:"info,omitempty"`
	SyncStatus *SyncStatus     `json:"sync-status,omitempty"`

	Satellites *Satellites `json:"satellites,omitempty"`
	GRC        *GRC        `json:"grc,omitempty"`

	DCF77 *DCF77 `json:"dcf77,omitempty"`
}

type SlotModuleInfo struct {
	Model            string       `json:"model"`
	SerialNumber     SerialNumber `json:"serial-number"`
	SoftwareRevision string       `json:"software-revision"`
	FirmwareImage    string       `json:"firmware-image"`
}

type SyncStatus struct {
	OscillatorType string       `json:"osc-type"`
	TimeQuality    *TimeQuality `json:"est-time-quality"`
	ClockStatus    ClockStatus  `json:"clock-status"`
}

type TimeQuality time.Duration

func (t TimeQuality) Seconds() float64 {
	return time.Duration(t).Seconds()
}

func (t *TimeQuality) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal time quality: %w", err)
	}

	trimmed := strings.TrimPrefix(strings.ToLower(raw), "less-than-")

	d, err := time.ParseDuration(trimmed)
	if err != nil {
		return fmt.Errorf("failed to parse time quality %q: %w", raw, err)
	}

	*t = TimeQuality(d)

	return nil
}

type ClockStatus struct {
	Clock      string `json:"clock"`
	Oscillator string `json:"oscillator"`
}

func (cs ClockStatus) IsSynchronized() bool {
	return cs.Clock == "synchronized"
}

func (cs ClockStatus) IsOscillatorWarmedUp() bool {
	return cs.Oscillator == "warmed-up"
}

type Satellites struct {
	InView    float64 `json:"satellites-in-view"`
	Good      float64 `json:"good-satellites"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type GRC struct {
	Antenna  *Antenna  `json:"antenna,omitempty"`
	Receiver *Receiver `json:"receiver,omitempty"`
}

type Antenna struct {
	IsConnected     bool `json:"connected"`
	HasShortCircuit bool `json:"short-circuit"`
}

type Receiver struct {
	IsSynchronized bool `json:"synchronized"`
	IsTracking     bool `json:"tracking"`
	IsColdBooting  bool `json:"cold-boot"`
	IsWarmBooting  bool `json:"warm-boot"`
}

type DCF77 struct {
	ID            string  `json:"type"`
	Name          string  `json:"ref-type"`
	Correlation   float64 `json:"correlation"`
	FieldStrength float64 `json:"field-strength"`
}
