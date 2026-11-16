package models

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestLeapIndicator_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  LeapIndicator
		expectErr bool
	}{
		{"int 0", `0`, LeapNoWarning, false},
		{"int 1", `1`, LeapLastMinuteHas61Seconds, false},
		{"int 2", `2`, LeapLastMinuteHas59Seconds, false},
		{"int 3", `3`, LeapUnknown, false},
		{"string 0", `"0"`, LeapNoWarning, false},
		{"string 3", `"3"`, LeapUnknown, false},
		{"int out of range", `4`, 0, true},
		{"int negative", `-1`, 0, true},
		{"string out of range", `"5"`, 0, true},
		{"string not a number", `"abc"`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var li LeapIndicator
			err := json.Unmarshal([]byte(tt.input), &li)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if li != tt.expected {
				t.Errorf("got %d, want %d", li, tt.expected)
			}
		})
	}
}

func TestUnixFromYYYYMMDDhhmm_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		expectErr bool
	}{
		{"valid", `"202512310000"`, float64(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC).Unix()), false},
		{"valid with time", `"202507010630"`, float64(time.Date(2025, 7, 1, 6, 30, 0, 0, time.UTC).Unix()), false},
		{"invalid format", `"not-a-date"`, 0, true},
		{"not a string", `123`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u UnixFromYYYYMMDDhhmm
			err := json.Unmarshal([]byte(tt.input), &u)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if float64(u) != tt.expected {
				t.Errorf("got %f, want %f", float64(u), tt.expected)
			}
		})
	}
}

func TestNTPAssociation_IsSys(t *testing.T) {
	if !(NTPAssociation{AssociationID: 0}).IsSys() {
		t.Error("expected IsSys() == true for AssociationID 0")
	}
	if (NTPAssociation{AssociationID: 1}).IsSys() {
		t.Error("expected IsSys() == false for AssociationID 1")
	}
}

func TestNTPAssociation_PrecisionSeconds(t *testing.T) {
	a := NTPAssociation{Precision: -20}
	got := a.PrecisionSeconds()
	want := math.Pow(2, -20)
	if got != want {
		t.Errorf("got %e, want %e", got, want)
	}
}
