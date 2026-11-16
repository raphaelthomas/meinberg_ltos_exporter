package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SerialNumber string

func (s *SerialNumber) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal serial number: %w", err)
	}

	sn := strings.TrimSpace(raw)
	switch strings.ToLower(sn) {
	case "", "unknown", "n/a", "na", "none":
		*s = ""
	default:
		*s = SerialNumber(sn)
	}

	return nil
}

func (s SerialNumber) String() string {
	return string(s)
}
