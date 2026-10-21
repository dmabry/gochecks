package main

import (
	"strings"
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

// TestDetermineInterfaceUsage_CriticalNotMaskedByWarning verifies that an
// outbound critical is not masked by an inbound warning: critical thresholds
// for both directions are evaluated before any warning threshold.
// Regression: inbound conditions were checked first, so an inbound warning
// reported Warning even when outbound traffic exceeded its critical threshold.
func TestDetermineInterfaceUsage_CriticalNotMaskedByWarning(t *testing.T) {
	// 750 Kbps inbound (exceeds the 500 Kbps warning, below the 1 Mbps
	// critical) and 3 Mbps outbound (exceeds the 2 Mbps critical): the
	// result must be Critical naming the outbound direction.
	first := testMetrics(baseTime(), 0)
	first.In = 0
	second := testMetrics(baseTime().Add(10*time.Second), 0)
	second.In = 93750 * 10   // 750 Kbps inbound rate
	second.Out = 375000 * 10 // 3 Mbps outbound rate
	second.HCIn = 93750 * 10
	second.HCOut = 375000 * 10

	got := DetermineInterfaceUsage(first, second, 500000, 1000000, 1000000, 2000000, false)

	if got.ExitCode != gomonitor.Critical {
		t.Errorf("got %s (%q), want Critical", got.ExitCode, got.Message)
	}
	if !strings.Contains(got.Message, "Outbound exceeds threshold") {
		t.Errorf("message %q should name the outbound direction", got.Message)
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

// TestCounterDelta32 verifies 32-bit counter wrap handling.
func TestCounterDelta32(t *testing.T) {
	tests := []struct {
		name     string
		previous uint
		current  uint
		want     uint64
	}{
		{name: "monotonic", previous: 100, current: 250, want: 150},
		{name: "no change", previous: 100, current: 100, want: 0},
		{name: "wrap", previous: 0xFFFFFF00, current: 0x100, want: 0x200},
		{name: "wrap at max", previous: 0xFFFFFFFF, current: 0, want: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := counterDelta32(tc.previous, tc.current); got != tc.want {
				t.Errorf("counterDelta32(%d, %d) = %d, want %d", tc.previous, tc.current, got, tc.want)
			}
		})
	}
}

// TestCounterDelta64 verifies that a decreasing 64-bit counter is reported as
// a reset rather than underflowing.
func TestCounterDelta64(t *testing.T) {
	if got, ok := counterDelta64(100, 250); !ok || got != 150 {
		t.Errorf("counterDelta64(100, 250) = (%d, %v), want (150, true)", got, ok)
	}
	if _, ok := counterDelta64(250, 100); ok {
		t.Error("counterDelta64(250, 100) reported a monotonic increase, want reset")
	}
}

// TestDetermineInterfaceUsage_32BitCounterWrap verifies a wrapped 32-bit
// counter does not produce a bogus rate. Regression: unsigned subtraction
// underflowed to ~2^64 and the message reported Gbps-scale garbage.
func TestDetermineInterfaceUsage_32BitCounterWrap(t *testing.T) {
	first := InterfaceMetrics{
		Name:      "eth0",
		In:        0xFFFFFF00,
		Out:       0xFFFFFF00,
		HCIn:      1_000_000,
		HCOut:     1_000_000,
		Speed:     1_000_000_000,
		Latency:   10 * time.Millisecond,
		Timestamp: baseTime(),
	}
	second := InterfaceMetrics{
		Name:      "eth0",
		In:        0x100, // wrapped: 512 octets
		Out:       0x100,
		HCIn:      1_000_512,
		HCOut:     1_000_512,
		Speed:     1_000_000_000,
		Latency:   10 * time.Millisecond,
		Timestamp: baseTime().Add(10 * time.Second),
	}

	got := DetermineInterfaceUsage(first, second, 1000, 1000, 2000, 2000, false)

	if got.ExitCode != gomonitor.OK {
		t.Errorf("got %s (%q), want OK", got.ExitCode, got.Message)
	}
	if strings.Contains(got.Message, "Gbps") {
		t.Errorf("message %q reports a wrapped counter as Gbps-scale traffic", got.Message)
	}
	if !strings.Contains(got.Message, "In: 408.0 bps") {
		t.Errorf("message %q does not report the wrapped 32-bit rate (~408 bps)", got.Message)
	}
}

// TestDetermineInterfaceUsage_HCCounterReset verifies a decreasing 64-bit
// counter returns Unknown instead of computing a rate across the reset.
func TestDetermineInterfaceUsage_HCCounterReset(t *testing.T) {
	first := testMetrics(baseTime(), 1000)
	second := testMetrics(baseTime().Add(10*time.Second), 2000)
	second.HCIn = first.HCIn - 1

	got := DetermineInterfaceUsage(first, second, 0, 0, 0, 0, false)

	if got.ExitCode != gomonitor.Unknown {
		t.Errorf("got %s (%q), want Unknown", got.ExitCode, got.Message)
	}
}

// TestDetermineInterfaceUsage_ZeroThresholdsDisabled verifies the default
// zero thresholds disable alerting instead of flagging any nonzero traffic.
// Regression: 0 was compared with >, so a default invocation returned Critical.
func TestDetermineInterfaceUsage_ZeroThresholdsDisabled(t *testing.T) {
	first := testMetrics(baseTime(), 0)
	second := testMetrics(baseTime().Add(10*time.Second), 1000) // nonzero traffic

	got := DetermineInterfaceUsage(first, second, 0, 0, 0, 0, false)

	if got.ExitCode != gomonitor.OK {
		t.Errorf("got %s (%q), want OK with thresholds disabled", got.ExitCode, got.Message)
	}
}

// TestDetermineInterfaceUsage_DisabledThresholdsOmitted verifies disabled
// thresholds are omitted from perfdata rather than emitted as zero (which
// Nagios/Icinga interpret as alert-on-any-nonzero).
func TestDetermineInterfaceUsage_DisabledThresholdsOmitted(t *testing.T) {
	first := testMetrics(baseTime(), 0)
	second := testMetrics(baseTime().Add(10*time.Second), 1000)

	got := DetermineInterfaceUsage(first, second, 0, 0, 0, 0, true)

	for _, name := range []string{"in", "out", "hc_in", "hc_out"} {
		metric, ok := got.PerformanceData[name]
		if !ok {
			t.Fatalf("missing %q performance metric", name)
		}
		if metric.Warn != nil || metric.Crit != nil {
			t.Errorf("%q metric has disabled thresholds: Warn=%v Crit=%v", name, metric.Warn, metric.Crit)
		}
	}
}
