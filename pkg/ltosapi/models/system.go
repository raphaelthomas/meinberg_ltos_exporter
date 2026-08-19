package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type System struct {
	UptimeSeconds float64  `json:"uptime"`
	CPULoad       CPULoad  `json:"cpuload"`
	Memory        Memory   `json:"memory"`
	Mounts        []Mount  `json:"storage"`
	Firmware      Firmware `json:"firmware"`
}

type Firmware struct {
	Running  string `json:"running"`
	Selected string `json:"selected"`
	// Image is the running firmware image of raw form "firmware-7.10.008-x32"
	Image string `json:"fwimage"`
}

type CPULoad struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

var (
	memTotalRe = regexp.MustCompile(`(\d+)\s+kB\s+total`)
	memFreeRe  = regexp.MustCompile(`(\d+)\s+kB\s+free`)
)

// UnmarshalJSON CPULoad of raw form "0.48 0.66 0.57 2/99 25157"
func (c *CPULoad) UnmarshalJSON(data []byte) error {
	var rawCPULoadStr string
	if err := json.Unmarshal(data, &rawCPULoadStr); err != nil {
		return fmt.Errorf("failed to unmarshal CPU load string: %v", err)
	}

	parts := strings.Fields(rawCPULoadStr)
	if len(parts) < 3 {
		return fmt.Errorf("failed to parse CPU load string, expected at least 3 fields: %s", rawCPULoadStr)
	}

	var err error
	c.Load1, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("failed to parse 1-minute CPU load: %v", err)
	}
	c.Load5, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse 5-minute CPU load: %v", err)
	}
	c.Load15, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return fmt.Errorf("failed to parse 15-minute CPU load: %v", err)
	}

	return nil
}

type Memory struct {
	Total float64
	Free  float64
}

// UnmarshalJSON memory of raw form "228428 kB total memory, 161732 kB free (70 %)"
func (m *Memory) UnmarshalJSON(data []byte) error {
	var rawMemoryStr string
	if err := json.Unmarshal(data, &rawMemoryStr); err != nil {
		return fmt.Errorf("failed to unmarshal memory string: %v", err)
	}

	totalMatches := memTotalRe.FindStringSubmatch(rawMemoryStr)
	if len(totalMatches) < 2 {
		return fmt.Errorf("failed to parse total memory: %q", rawMemoryStr)
	}
	totalMemoryKB, err := strconv.ParseFloat(totalMatches[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse total memory as float: %v", err)
	}

	freeMatches := memFreeRe.FindStringSubmatch(rawMemoryStr)
	if len(freeMatches) < 2 {
		return fmt.Errorf("failed to parse free memory: %q", rawMemoryStr)
	}
	freeMemoryKB, err := strconv.ParseFloat(freeMatches[1], 64)
	if err != nil {
		return fmt.Errorf("failed to parse free memory as float: %v", err)
	}

	m.Total = totalMemoryKB * 1024
	m.Free = freeMemoryKB * 1024

	return nil
}

type Mount struct {
	Size       float64 `json:"size"`
	Used       float64 `json:"used"`
	Mountpoint string  `json:"mountpoint"`
}

func (m *Mount) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Size       float64 `json:"size"`
		Used       float64 `json:"used"`
		Mountpoint string  `json:"mountpoint"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal mount: %v", err)
	}

	m.Size = aux.Size * 1024
	m.Used = aux.Used * 1024
	m.Mountpoint = aux.Mountpoint

	return nil
}
