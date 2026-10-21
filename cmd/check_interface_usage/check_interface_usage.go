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
	"context"
	"flag"
	"fmt"
	"github.com/dmabry/gochecks/internal/interfaces"
	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/dmabry/gomonitor"
	"strconv"
	"time"

	"github.com/gosnmp/gosnmp"
)

// InterfaceMetrics represents the metrics of a network interface.
type InterfaceMetrics struct {
	Name      string
	In        uint
	Out       uint
	HCIn      uint64
	HCOut     uint64
	Speed     uint
	HighSpeed uint
	Latency   time.Duration
	Timestamp time.Time
}

type InterfaceStatus struct {
	AdminStatus int
	OperStatus  int
}

func (ifStatus *InterfaceStatus) IsInterfaceUp() bool {
	// If AdminStatus or OperStatus is not 1, return false
	if ifStatus.AdminStatus != 1 || ifStatus.OperStatus != 1 {
		return false
	}

	// If both are 1, return true
	return true
}

func GetInterfaceStatus(snmpClient *snmp.Client, index int) (*InterfaceStatus, error) {
	strIndex := strconv.Itoa(index)
	oidAdminStatus := fmt.Sprintf("%s.%s", interfaces.OIDIfAdminStatus, strIndex)
	oidOperStatus := fmt.Sprintf("%s.%s", interfaces.OIDIfOperStatus, strIndex)

	statusOIDs := []string{oidAdminStatus, oidOperStatus}

	result, _, err := snmpClient.GetValue(context.TODO(), statusOIDs)
	if err != nil {
		return nil, fmt.Errorf("requested OID: %w", err)
	}

	if len(result.Variables) < 2 || result.Variables[0].Value == nil || result.Variables[1].Value == nil {
		return nil, fmt.Errorf("interface index %d does not exist", index)
	}

	adminStatus, ok := result.Variables[0].Value.(int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for adminStatus: %T", result.Variables[0].Value)
	}
	operStatus, ok := result.Variables[1].Value.(int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for operStatus: %T", result.Variables[1].Value)
	}

	return &InterfaceStatus{
		AdminStatus: adminStatus,
		OperStatus:  operStatus,
	}, nil
}

// convertToScale converts a given value to the appropriate scale (bps, Kbps, Mbps, or Gbps).
// The function takes an input value in octets per second and returns the converted value
// along with the corresponding unit of measurement. Floating-point math is used so
// sub-unit precision (e.g. 12.9 Mbps) is preserved instead of truncated.
//
// Parameters:
//   - value: The input value in octets per second to be converted.
//
// Returns:
//   - out: The converted value in the appropriate scale (bps, Kbps, Mbps, or Gbps).
//   - unit: The corresponding unit of measurement for the converted value.
func convertToScale(value uint64) (out float64, unit string) {
	bps := float64(value) * 8
	switch {
	case bps < 1000:
		return bps, "bps"
	case bps < 1000000:
		return bps / 1000, "Kbps"
	case bps < 1000000000:
		return bps / 1000000, "Mbps"
	default:
		return bps / 1000000000, "Gbps"
	}
}

// GetInterfaceMetrics retrieves the network interface metrics for a specific interface
// using the provided SNMP client and index.
//
// Parameters:
//   - snmpClient: The SNMP client used to connect and retrieve the metrics.
//   - index: The index of the interface to retrieve the metrics for.
//
// Returns:
//   - metrics: The network interface metrics for the specified interface.
//   - error: Any error encountered during the retrieval of the metrics.
func GetInterfaceMetrics(snmpClient *snmp.Client, index int) (*InterfaceMetrics, error) {
	strIndex := strconv.Itoa(index)
	oidName := interfaces.OIDIfName + "." + strIndex
	oidHCIn := interfaces.OIDIfHCInOctets + "." + strIndex
	oidHCOut := interfaces.OIDIfHCOutOctets + "." + strIndex
	oidIn := interfaces.OIDIfInOctets + "." + strIndex
	oidOut := interfaces.OIDIfOutOctets + "." + strIndex
	oidSpeed := interfaces.OIDIfSpeed + "." + strIndex
	oidHighSpeed := interfaces.OIDIfHighSpeed + "." + strIndex
	usageOIDs := []string{oidName, oidIn, oidOut, oidHCIn, oidHCOut, oidSpeed, oidHighSpeed}

	result, latency, err := snmpClient.GetValue(context.TODO(), usageOIDs)
	if err != nil {
		return nil, fmt.Errorf("requested OID: %w", err)
	}

	if len(result.Variables) < len(usageOIDs) {
		return nil, fmt.Errorf("interface index %d: expected %d varbinds, got %d", index, len(usageOIDs), len(result.Variables))
	}

	name, err := varbindString(result.Variables[0].Value)
	if err != nil {
		return nil, fmt.Errorf("ifName: %w", err)
	}
	in, err := varbindUint(result.Variables[1].Value)
	if err != nil {
		return nil, fmt.Errorf("ifInOctets: %w", err)
	}
	out, err := varbindUint(result.Variables[2].Value)
	if err != nil {
		return nil, fmt.Errorf("ifOutOctets: %w", err)
	}
	hcIn, err := varbindUint64(result.Variables[3].Value)
	if err != nil {
		return nil, fmt.Errorf("ifHCInOctets: %w", err)
	}
	hcOut, err := varbindUint64(result.Variables[4].Value)
	if err != nil {
		return nil, fmt.Errorf("ifHCOutOctets: %w", err)
	}
	speed, err := varbindUint(result.Variables[5].Value)
	if err != nil {
		return nil, fmt.Errorf("ifSpeed: %w", err)
	}
	highSpeed, err := varbindUint(result.Variables[6].Value)
	if err != nil {
		return nil, fmt.Errorf("ifHighSpeed: %w", err)
	}

	metrics := &InterfaceMetrics{
		Name:      name,
		In:        in,
		Out:       out,
		HCIn:      hcIn,
		HCOut:     hcOut,
		Speed:     speed,
		HighSpeed: highSpeed,
		Latency:   latency,
		Timestamp: time.Now(),
	}

	return metrics, nil
}

// varbindString accepts the string shapes an SNMP octet string arrives in:
// gosnmp []byte or a plain Go string.
func varbindString(value interface{}) (string, error) {
	switch v := value.(type) {
	case []byte:
		return string(v), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("expected string value, got %T", value)
	}
}

// varbindUint accepts uint plus non-negative int and int64, covering
// Counter32 and Gauge32 values which gosnmp decodes as uint.
func varbindUint(value interface{}) (uint, error) {
	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("expected non-negative value, got %d", v)
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("expected non-negative value, got %d", v)
		}
		return uint(v), nil
	default:
		return 0, fmt.Errorf("expected uint value, got %T", value)
	}
}

// varbindUint64 accepts uint64 plus non-negative int and int64, covering
// Counter64 values which gosnmp decodes as uint64.
func varbindUint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case uint64:
		return v, nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("expected non-negative value, got %d", v)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("expected non-negative value, got %d", v)
		}
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("expected uint64 value, got %T", value)
	}
}

// counterDelta32 returns the octet increase between two samples of a 32-bit
// counter, accounting for wrap at 2^32. A decrease is treated as a wrap; a
// device reboot also resets the HC counters, which is detected separately.
func counterDelta32(previous, current uint) uint64 {
	if current >= previous {
		return uint64(current - previous)
	}
	return uint64(current) + (uint64(1) << 32) - uint64(previous)
}

// counterDelta64 returns the octet increase between two samples of a 64-bit
// counter and whether the samples are monotonic. A decrease indicates a
// counter reset or discontinuity, so the rate across it is undefined.
func counterDelta64(previous, current uint64) (uint64, bool) {
	if current >= previous {
		return current - previous, true
	}
	return 0, false
}

// exceeds reports whether value is above threshold. Thresholds at or below
// zero are disabled, which matches the flag defaults and avoids alerting on
// any nonzero traffic when no thresholds are supplied.
func exceeds(value uint64, threshold int) bool {
	return threshold > 0 && value > uint64(threshold)
}

// thresholdPtr returns a performance-data threshold pointer, or nil when the
// threshold is disabled so it is omitted from the perfdata output.
func thresholdPtr(threshold int) *float64 {
	if threshold <= 0 {
		return nil
	}
	return new(float64(threshold))
}

// DetermineInterfaceUsage calculates the usage of a network interface based on the provided InterfaceMetrics.
// It compares the metrics between two time periods and determines if the inbound and outbound traffic exceeds
// the given warning and critical thresholds. It also converts the traffic values to the appropriate scale (bps,
// Kbps, Mbps, or Gbps) and crafts a message with the results.
//
// Parameters:
//   - first: The InterfaceMetrics representing the metrics of the first time period.
//   - second: The InterfaceMetrics representing the metrics of the second time period.
//   - warnIn: The warning threshold for inbound traffic in bps.
//   - warnOut: The warning threshold for outbound traffic in bps.
//   - critIn: The critical threshold for inbound traffic in bps.
//   - critOut: The critical threshold for outbound traffic in bps.
//   - enablePerf: A boolean indicating whether to include performance data in the check result.
//
// Returns:
//   - checkResult: A pointer to a gomonitor.CheckResult object that represents the result of the interface usage calculation.
//
// Example:
//
//	first := InterfaceMetrics{Name: "eth0", In: 100, Out: 200, HCIn: 300, HCOut: 400, Speed: 1000, Latency: 10 * time.Millisecond, Timestamp: time.Now()}
//	second := InterfaceMetrics{Name: "eth0", In: 200, Out: 300, HCIn: 400, HCOut: 500, Speed: 1000, Latency: 20 * time.Millisecond, Timestamp: time.Now()}
//	result := DetermineInterfaceUsage(first, second, 500, 500, 1000, 1000, true)
//	result.SendResult()
func DetermineInterfaceUsage(first InterfaceMetrics, second InterfaceMetrics, warnIn int, warnOut int, critIn int, critOut int, enablePerf bool) *gomonitor.CheckResult {
	checkResult := gomonitor.NewCheckResult()
	intName := first.Name
	periodDiff := second.Timestamp.Sub(first.Timestamp)
	period := periodDiff.Seconds()

	// Guard against a non-positive measurement period: a zero or negative
	// period would divide by zero below, and a sub-second period would
	// truncate to 0 with uint(period). Return Unknown so Nagios reports a
	// plugin problem rather than crashing or reporting bogus rates.
	if period <= 0 {
		checkResult.SetResult(gomonitor.Unknown, fmt.Sprintf("measurement period is %.3f seconds; two measurements with distinct timestamps are required (check the -delay flag)", period))
		return checkResult
	}

	avgLatency := (first.Latency + second.Latency) / 2

	// Counters can wrap (32-bit counters wrap frequently on high-speed links)
	// or reset (device reboot, counter clear). Compute deltas with wrap
	// handling for 32-bit counters; a decreasing 64-bit counter is a reset,
	// so the rate across it is undefined.
	inDelta := counterDelta32(first.In, second.In)
	outDelta := counterDelta32(first.Out, second.Out)
	hcInDelta, hcInMonotonic := counterDelta64(first.HCIn, second.HCIn)
	hcOutDelta, hcOutMonotonic := counterDelta64(first.HCOut, second.HCOut)
	if !hcInMonotonic || !hcOutMonotonic {
		checkResult.SetResult(gomonitor.Unknown, fmt.Sprintf(
			"counter reset or discontinuity between measurements (HCIn %d -> %d, HCOut %d -> %d); unable to compute a rate",
			first.HCIn, second.HCIn, first.HCOut, second.HCOut))
		return checkResult
	}

	// Calc rates. Floating-point division is used so sub-second measurement
	// periods work correctly; integer division with uint(period) would
	// truncate a 0.5s period to 0 and divide by zero.
	inRate := float64(inDelta) / period
	outRate := float64(outDelta) / period
	hcInRate := float64(hcInDelta) / period
	hcOutRate := float64(hcOutDelta) / period
	in := uint64(inRate)
	out := uint64(outRate)
	hcIn := uint64(hcInRate)
	hcOut := uint64(hcOutRate)
	// Convert to scale for the human-readable message
	intIn, intInUnit := convertToScale(in)
	intOut, intOutUnit := convertToScale(out)
	intHCIn, intHCInUnit := convertToScale(hcIn)
	intHCOut, intHCOutUnit := convertToScale(hcOut)
	// Craft message
	message := fmt.Sprintf("%s - In: %.1f %s Out: %.1f %s HCIn: %.1f %s HCOut: %.1f %s", intName, intIn, intInUnit, intOut, intOutUnit, intHCIn, intHCInUnit, intHCOut, intHCOutUnit)
	if enablePerf {
		checkResult.AddPerformanceData("snmp_latency", gomonitor.PerformanceMetric{Value: avgLatency.Seconds(), UnitOM: "s"})
		checkResult.AddPerformanceData("in", gomonitor.PerformanceMetric{Value: inRate * 8, Warn: thresholdPtr(warnIn), Crit: thresholdPtr(critIn), Min: new(0.0), Max: new(float64(first.Speed)), UnitOM: "bps"})
		checkResult.AddPerformanceData("out", gomonitor.PerformanceMetric{Value: outRate * 8, Warn: thresholdPtr(warnOut), Crit: thresholdPtr(critOut), Min: new(0.0), Max: new(float64(first.Speed)), UnitOM: "bps"})
		checkResult.AddPerformanceData("hc_in", gomonitor.PerformanceMetric{Value: hcInRate * 8, Warn: thresholdPtr(warnIn), Crit: thresholdPtr(critIn), Min: new(0.0), Max: new(float64(first.Speed)), UnitOM: "bps"})
		checkResult.AddPerformanceData("hc_out", gomonitor.PerformanceMetric{Value: hcOutRate * 8, Warn: thresholdPtr(warnOut), Crit: thresholdPtr(critOut), Min: new(0.0), Max: new(float64(first.Speed)), UnitOM: "bps"})
	}

	// Compare in bps: warnIn/critIn/warnOut/critOut are documented in bps, so
	// the thresholds are matched against the raw bits-per-second rates
	// (octets * 8), not the human-scaled values (Kbps/Mbps/Gbps) which only
	// belong in the message. Comparing scaled values against bps thresholds
	// silently disabled alerting.
	inBps := uint64(inRate * 8)
	outBps := uint64(outRate * 8)
	hcInBps := uint64(hcInRate * 8)
	hcOutBps := uint64(hcOutRate * 8)

	// Prefer the 64-bit HC counters for threshold evaluation: they do not
	// wrap in practice and avoid the 32-bit wrap-versus-reset ambiguity. Fall
	// back to the wrap-corrected 32-bit counters when the device reports no
	// HC counter activity (some agents return 0 for unsupported columns).
	thresholdInBps, thresholdOutBps := hcInBps, hcOutBps
	if first.HCIn == 0 && second.HCIn == 0 && first.HCOut == 0 && second.HCOut == 0 {
		thresholdInBps, thresholdOutBps = inBps, outBps
	}

	// Evaluate critical thresholds for both directions before warnings so an
	// inbound warning cannot mask an outbound critical. A zero threshold (the
	// flag default) disables that threshold.
	switch {
	case exceeds(thresholdInBps, critIn):
		checkResult.SetResult(gomonitor.Critical, "Inbound exceeds threshold "+message)
	case exceeds(thresholdOutBps, critOut):
		checkResult.SetResult(gomonitor.Critical, "Outbound exceeds threshold "+message)
	case exceeds(thresholdInBps, warnIn):
		checkResult.SetResult(gomonitor.Warning, "Inbound exceeds threshold "+message)
	case exceeds(thresholdOutBps, warnOut):
		checkResult.SetResult(gomonitor.Warning, "Outbound exceeds threshold "+message)
	default:
		checkResult.SetResult(gomonitor.OK, message)
	}
	return checkResult
}

func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	index := flag.Int("index", 1, "The index of the Interface")
	delay := flag.Int("delay", 10, "The delay in seconds to wait between measurements")
	enablePerfData := flag.Bool("enablePerfData", false, "Enable performance data. Default is false.")
	warnIn := flag.Int("warnIn", 0, "Warning level for inbound in bps. 0 disables the threshold.")
	critIn := flag.Int("critIn", 0, "Critical level for inbound in bps. 0 disables the threshold.")
	warnOut := flag.Int("warnOut", 0, "Warning level for outbound in bps. 0 disables the threshold.")
	critOut := flag.Int("critOut", 0, "Critical level for outbound bps. 0 disables the threshold.")
	checkStatus := flag.Bool("checkStatus", true, "Check interface admin/oper status before measuring. Returns Critical if the interface is down.")
	snmpVersion := &snmp.SNMPVersionFlag{}
	flag.Var(snmpVersion, "snmpVersion", "SNMP protocol version: 2c or 3 (default 2c).")
	v3Username, v3AuthProtocol, v3AuthPassphrase, v3PrivProtocol, v3PrivPassphrase := snmp.AddV3Flags(flag.CommandLine)
	flag.Parse()

	version := snmpVersion.VersionOrDefault()

	snmpClient, err := snmp.NewClient(*target, *community, snmp.WithSNMPVersion(version))
	if err != nil {
		checkResult := gomonitor.NewCheckResult()
		checkResult.SetResult(gomonitor.Unknown, fmt.Sprintf("failed to configure SNMP client: %v", err))
		checkResult.SendResult()
		return
	}

	if version == gosnmp.Version3 {
		if err := snmp.ApplyV3Flags(snmpClient, *v3Username, v3AuthProtocol.Protocol, *v3AuthPassphrase, v3PrivProtocol.Protocol, *v3PrivPassphrase); err != nil {
			checkResult := gomonitor.NewCheckResult()
			checkResult.SetResult(gomonitor.Unknown, err.Error())
			checkResult.SendResult()
			return
		}
	}

	// Determine Interface Status before proceeding
	if *checkStatus {
		status, err := GetInterfaceStatus(snmpClient, *index)
		if err != nil {
			checkResult := gomonitor.NewCheckResult()
			eMessage := fmt.Sprintf("SNMP target %s failed to return data when getting interface status. %s", snmpClient.Target, err)
			checkResult.SetResult(gomonitor.Critical, eMessage)
			checkResult.SendResult()
			return
		}

		if !status.IsInterfaceUp() {
			checkResult := gomonitor.NewCheckResult()
			eMessage := fmt.Sprintf("SNMP target %s interface %d is down (adminStatus=%d, operStatus=%d).", snmpClient.Target, *index, status.AdminStatus, status.OperStatus)
			checkResult.SetResult(gomonitor.Critical, eMessage)
			checkResult.SendResult()
			return
		}
	}

	measure1, err1 := GetInterfaceMetrics(snmpClient, *index)
	if err1 != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return data when measuring metrics. %s", snmpClient.Target, err1)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		checkResult.SendResult()
		return
	}

	// delay
	time.Sleep(time.Duration(*delay) * time.Second)

	measure2, err2 := GetInterfaceMetrics(snmpClient, *index)
	if err2 != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return data when measuring metrics. %s", snmpClient.Target, err2)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		checkResult.SendResult()
		return
	}

	// Calculate current usage and determine thresholds
	result := DetermineInterfaceUsage(*measure1, *measure2, *warnIn, *warnOut, *critIn, *critOut, *enablePerfData)
	result.SendResult()
}
