/*
   Copyright 2024 David Mabry

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"testing"
	"time"

	"github.com/dmabry/gochecks/internal/snmp"
)

func TestConvertToScale(t *testing.T) {
	tests := []struct {
		name      string
		input     uint64
		wantValue uint64
		wantUnit  string
	}{
		{name: "zero", input: 0, wantValue: 0, wantUnit: "bps"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit := convertToScale(tt.input)
			if value != tt.wantValue || unit != tt.wantUnit {
				t.Errorf("convertToScale(%d) = (%d, %s), want (%d, %s)", tt.input, value, unit, tt.wantValue, tt.wantUnit)
			}
		})
	}
}

func TestConvertToScaleBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		input     uint64
		wantValue uint64
		wantUnit  string
	}{
		{name: "1 byte", input: 1, wantValue: 8, wantUnit: "bps"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit := convertToScale(tt.input)
			if value != tt.wantValue || unit != tt.wantUnit {
				t.Errorf("convertToScale(%d) = (%d, %s), want (%d, %s)", tt.input, value, unit, tt.wantValue, tt.wantUnit)
			}
		})
	}
}

func TestConvertToScaleOverflow(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		wantUnit string
	}{
		{name: "max uint64", input: 18446744073709551615, wantUnit: "Gbps"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit := convertToScale(tt.input)
			if unit != tt.wantUnit {
				t.Errorf("convertToScale() returned unit %s, want %s", unit, tt.wantUnit)
			}
			_ = value
		})
	}
}

func TestConvertToScaleKbpsBoundary(t *testing.T) {
	tests := []struct {
		name      string
		input     uint64
		wantValue uint64
		wantUnit  string
	}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit := convertToScale(tt.input)
			if value != tt.wantValue || unit != tt.wantUnit {
				t.Errorf("convertToScale(%d) = (%d, %s), want (%d, %s)", tt.input, value, unit, tt.wantValue, tt.wantUnit)
			}
		})
	}
}

func TestDetermineInterfaceUsage(t *testing.T) {
	result := DetermineInterfaceUsage(
		testMetrics(baseTime(), 100),
		testMetrics(baseTime().Add(time.Second), 200),
		0, 0, 0, 0,
		false,
	)

	if result == nil {
		t.Fatal("DetermineInterfaceUsage() returned nil")
	}
}

func baseTime() time.Time {
	return time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
}

func testMetrics(ts time.Time, in uint) InterfaceMetrics {
	return InterfaceMetrics{
		Name:      "eth0",
		In:        in,
		Out:       in * 2,
		HCIn:      uint64(in * 3),
		HCOut:     uint64(in * 4),
		Speed:     1000000000,
		HighSpeed: 1000,
		Latency:   10 * time.Millisecond,
		Timestamp: ts,
	}
}

func TestGetInterfaceMetricsRequiresDevice(t *testing.T) {
	t.Skip("This test requires a real SNMP device")
	client := &snmp.Client{Target: "127.0.0.1", Community: "public"}
	_, _ = GetInterfaceMetrics(client, 1)
}

func TestIsInterfaceUp(t *testing.T) {
	tests := []struct {
		name        string
		adminStatus int
		operStatus  int
		want        bool
	}{
		{name: "both up", adminStatus: 1, operStatus: 1, want: true},
		{name: "admin down", adminStatus: 2, operStatus: 1, want: false},
		{name: "oper down", adminStatus: 1, operStatus: 2, want: false},
		{name: "both down", adminStatus: 2, operStatus: 2, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InterfaceStatus{AdminStatus: tt.adminStatus, OperStatus: tt.operStatus}
			if got := s.IsInterfaceUp(); got != tt.want {
				t.Errorf("IsInterfaceUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInterfaceStatusRequiresDevice(t *testing.T) {
	t.Skip("This test requires a real SNMP device")
	client := &snmp.Client{Target: "127.0.0.1", Community: "public"}
	_, _ = GetInterfaceStatus(client, 1)
}
