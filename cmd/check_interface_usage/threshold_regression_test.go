package main

import (
	"testing"
	"time"

	"github.com/dmabry/gomonitor"
)

// Regression tests for the threshold unit mismatch and the zero-period
// division by zero in DetermineInterfaceUsage.

// TestDetermineInterfaceUsage_ThresholdsInBps verifies that warnIn/critIn/
// warnOut/critOut — documented in bps — are matched against raw bits-per-second
// rates, not the human-scaled values (Kbps/Mbps/Gbps) used in the message.
// Regression: thresholds were compared against scaled values, so a 2 Mbps
// stream never tripped a 1 Mbps (1,000,000 bps) critical threshold.
func TestDetermineInterfaceUsage_ThresholdsInBps(t *testing.T) {
	tests := []struct {
		name     string
		in       uint // octets per period
		out      uint // octets per period
		hcIn     uint64
		hcOut    uint64
		warnIn   int
		critIn   int
		warnOut  int
		critOut  int
		wantCode gomonitor.ExitCode
	}{
		{
			// 2 Mbps inbound, 1 Mbps critical threshold: must be Critical
			name: "2Mbps exceeds 1Mbps critical", in: 250000, out: 0, hcIn: 250000, hcOut: 0,
			warnIn: 500000, critIn: 1000000, warnOut: 0, critOut: 0,
			wantCode: gomonitor.Critical,
		},
		{
			// 600 Kbps inbound, 1 Mbps critical / 500 Kbps warning: Warning
			name: "600Kbps exceeds 500Kbps warning", in: 75000, out: 0, hcIn: 75000, hcOut: 0,
			warnIn: 500000, critIn: 1000000, warnOut: 0, critOut: 0,
			wantCode: gomonitor.Warning,
		},
		{
			// 400 Kbps inbound, below both thresholds: OK
			name: "400Kbps below thresholds", in: 50000, out: 0, hcIn: 50000, hcOut: 0,
			warnIn: 500000, critIn: 1000000, warnOut: 0, critOut: 0,
			wantCode: gomonitor.OK,
		},
		{
			// Outbound thresholds use out/hcOut counters: 3 Mbps outbound
			// with a 2 Mbps critical threshold must be Critical
			name: "3Mbps outbound exceeds 2Mbps critical", in: 0, out: 375000, hcIn: 0, hcOut: 375000,
			warnIn: 0, critIn: 0, warnOut: 1000000, critOut: 2000000,
			wantCode: gomonitor.Critical,
		},
		{
			// HC counters alone can trip the threshold (non-hc counters wrap)
			name: "hcIn trips critical while in stays low", in: 0, out: 0, hcIn: 625000, hcOut: 0,
			warnIn: 1000000, critIn: 4000000, warnOut: 0, critOut: 0,
			wantCode: gomonitor.Critical, // 625000 octets/s = 5 Mbps > 4 Mbps
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			first := testMetrics(baseTime(), 0)
			first.In = 0
			second := testMetrics(baseTime().Add(10*time.Second), 0)
			// 10-second period: octets counter grows by tc.in * 10
			second.In = uint(tc.in) * 10
			second.Out = tc.out * 10
			second.HCIn = tc.hcIn * 10
			second.HCOut = tc.hcOut * 10

			got := DetermineInterfaceUsage(first, second, tc.warnIn, tc.warnOut, tc.critIn, tc.critOut, false)

			if got.ExitCode != tc.wantCode {
				t.Errorf("got %s (%q), want %s", got.ExitCode, got.Message, tc.wantCode)
			}
		})
	}
}

// TestDetermineInterfaceUsage_ZeroPeriod verifies a non-positive measurement
// period returns Unknown with a clear message instead of panicking with
// "integer divide by zero".
func TestDetermineInterfaceUsage_ZeroPeriod(t *testing.T) {
	first := testMetrics(baseTime(), 100)
	second := testMetrics(baseTime(), 200) // identical timestamps: period 0

	got := DetermineInterfaceUsage(first, second, 10, 10, 20, 20, false)

	if got.ExitCode != gomonitor.Unknown {
		t.Errorf("got %s (%q), want Unknown", got.ExitCode, got.Message)
	}
}

// TestDetermineInterfaceUsage_SubSecondPeriod verifies a period shorter than
// one second is not truncated to zero by uint(period) — it returns a real
// result instead of dividing by zero.
func TestDetermineInterfaceUsage_SubSecondPeriod(t *testing.T) {
	first := testMetrics(baseTime(), 0)
	second := testMetrics(baseTime().Add(500*time.Millisecond), 1000) // 0.5s period

	got := DetermineInterfaceUsage(first, second, 100000, 100000, 200000, 200000, false)

	if got.ExitCode == gomonitor.Unknown {
		t.Errorf("sub-second period was rejected as Unknown: %q (should compute rates)", got.Message)
	}
}
