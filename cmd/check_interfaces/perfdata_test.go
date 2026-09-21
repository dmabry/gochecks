package main

import (
	"strings"
	"testing"

	"github.com/dmabry/gomonitor"
)

// TestCountInterfaces verifies interface counting from a rendered
// interface-details message.
func TestCountInterfaces(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    int
	}{
		{name: "empty message", message: "", want: 0},
		{name: "single interface", message: "Interface index: 1\nDescription: eth0\nAlias: \n", want: 1},
		{
			name: "three interfaces",
			message: "Interface index: 1\nDescription: eth0\n\n" +
				"Interface index: 2\nDescription: eth1\n\n" +
				"Interface index: 3\nDescription: eth2",
			want: 3,
		},
		{name: "whitespace only", message: "  \n\n  ", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := countInterfaces(tc.message); got != tc.want {
				t.Errorf("countInterfaces(%q) = %d, want %d", tc.message, got, tc.want)
			}
		})
	}
}

// TestEnablePerfDataFlagWiring verifies that the enablePerfData flag adds an
// 'interfaces' performance metric to the check result (counting from a
// rendered message) and that the metric appears in FormatResult output.
func TestEnablePerfDataFlagWiring(t *testing.T) {
	// Simulate what main does with enablePerfData: count interfaces from the
	// rendered message and add the metric.
	message := "Interface index: 1\nDescription: eth0\n\n" +
		"Interface index: 2\nDescription: eth1"

	ifaceCount := countInterfaces(message)
	if ifaceCount != 2 {
		t.Fatalf("countInterfaces = %d, want 2", ifaceCount)
	}

	result := gomonitor.NewCheckResult()
	result.SetResult(gomonitor.OK, message)
	result.AddPerformanceData("interfaces", gomonitor.PerformanceMetric{Value: float64(ifaceCount)})

	got := result.FormatResult()
	if !strings.Contains(got, "'interfaces'=2.00") {
		t.Errorf("FormatResult %q does not contain the interfaces metric", got)
	}
}
