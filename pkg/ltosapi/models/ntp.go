package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

type LeapIndicator int

// See RFC 5905
const (
	NoWarning              LeapIndicator = iota // no warning
	LastMinuteHas61Seconds                      // last minute of the day has 61 seconds
	LastMinuteHas59Seconds                      // last minute of the day has 59 seconds
	Unknown                                     // unknown (clock unsynchronized)
)

func (l *LeapIndicator) setFromInt(n int) error {
	if n < int(NoWarning) || n > int(Unknown) {
		return fmt.Errorf("invalid leap indicator value: %d", n)
	}
	*l = LeapIndicator(n)
	return nil
}

func (l *LeapIndicator) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		return l.setFromInt(n)
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("failed to unmarshal leap indicator as string: %w", err)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid leap indicator string value: %q", s)
	}

	return l.setFromInt(n)
}

type UnixFromYYYYMMDDhhmm float64

func (u *UnixFromYYYYMMDDhhmm) UnmarshalJSON(data []byte) error {
	var leapSecRaw string
	if err := json.Unmarshal(data, &leapSecRaw); err != nil {
		return fmt.Errorf("failed to unmarshal leap second string: %w", err)
	}
	leapSecTime, err := time.Parse("200601021504", leapSecRaw)
	if err != nil {
		return fmt.Errorf("failed to parse leap second time: %w", err)
	}

	*u = UnixFromYYYYMMDDhhmm(leapSecTime.Unix())
	return nil
}

type NTPAssociation struct {
	AssociationID  int                  `json:"association-id"`
	Address        string               `json:"id"`
	Name           string               `json:"object-id"`
	RefID          string               `json:"refid"`
	Stratum        float64              `json:"stratum"`
	Precision      float64              `json:"precision"`
	RootDelay      float64              `json:"rootdelay"`
	RootDispersion float64              `json:"rootdisp"`
	ClockWander    float64              `json:"clk-wander"`
	ClockJitter    float64              `json:"clk-jitter"`
	LeapIndicator  LeapIndicator        `json:"leap"`
	LeapSecondUnix UnixFromYYYYMMDDhhmm `json:"leapsec"`

	// optional peer-only metrics
	Offset     *float64 `json:"offset,omitempty"`
	Delay      *float64 `json:"delay,omitempty"`
	Dispersion *float64 `json:"dispersion,omitempty"`
}

func (a NTPAssociation) IsSys() bool {
	return a.AssociationID == 0
}

func (a NTPAssociation) PrecisionSeconds() float64 {
	return math.Pow(2, a.Precision)
}
