package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Notification struct {
	Events []Event `json:"events"`
}

type Event struct {
	Type              string
	Name              string
	LastTriggeredUnix float64
}

func (e *Event) UnmarshalJSON(data []byte) error {
	aux := struct {
		Type          string `json:"type"`
		Name          string `json:"object-id"`
		LastTriggered string `json:"last-triggered"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	e.Type = aux.Type
	e.Name = aux.Name

	if aux.LastTriggered != "never" {
		// time.Parse without a timezone defaults to UTC. The Meinberg LTOS API returns timestamps without timezone information, and is assumed to use UTC.
		parsedTime, err := time.Parse("2006-01-02T15:04:05", aux.LastTriggered)
		if err != nil {
			return fmt.Errorf("failed to parse last-triggered timestamp %q: %w", aux.LastTriggered, err)
		}
		e.LastTriggeredUnix = float64(parsedTime.Unix())
	}

	return nil
}
