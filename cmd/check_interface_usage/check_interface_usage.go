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
	"flag"
	"fmt"
	"github.com/dmabry/gochecks/internal/interfaces"
	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/dmabry/gomonitor"
	"strconv"
	"time"
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

// convertToScale converts a given value to the appropriate scale (bps, Kbps, Mbps, or Gbps).
// The function takes an input value in bits per second (bps) and returns the converted value
// along with the corresponding unit of measurement.
//
// Parameters:
//   - value: The input value in bits per second (bps) to be converted.
//
// Returns:
//   - out: The converted value in the appropriate scale (bps, Kbps, Mbps, or Gbps).
//   - unit: The corresponding unit of measurement for the converted value.
func convertToScale(value uint64) (out uint64, unit string) {
	bps := value * 8
	if bps < 1000 {
		return bps, "bps"
	}

	kbps := bps / 1000 // convert octets to Kbps
	if kbps < 1000 {
		return kbps, "Kbps"
	}

	mbps := kbps / 1000 // convert Kbps to Mbps
	if mbps < 1000 {
		return mbps, "Mbps"
	}

	gbps := mbps / 1000 // convert Mbps to Gbps
	return gbps, "Gbps"
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
	oidName := fmt.Sprintf("%s.%s", interfaces.OIDIfName, strIndex)
	oidHCIn := fmt.Sprintf("%s.%s", interfaces.OIDIfHCInOctets, strIndex)
	oidHCOut := fmt.Sprintf("%s.%s", interfaces.OIDIfHCOutOctets, strIndex)
	oidIn := fmt.Sprintf("%s.%s", interfaces.OIDIfInOctets, strIndex)
	oidOut := fmt.Sprintf("%s.%s", interfaces.OIDIfOutOctets, strIndex)
	oidSpeed := fmt.Sprintf("%s.%s", interfaces.OIDIfSpeed, strIndex)
	oidHighSpeed := fmt.Sprintf("%s.%s", interfaces.OIDIfHighSpeed, strIndex)
	usageOIDs := []string{oidName, oidIn, oidOut, oidHCIn, oidHCOut, oidSpeed, oidHighSpeed}

	result, latency, err := snmpClient.GetValue(usageOIDs)
	if err != nil {
		eMessage := fmt.Sprintf("Requested OID: %s", err)
		return nil, fmt.Errorf("%s: %w", eMessage, err)
	}

	if result.Variables[0].Value == nil {
		eMessage := fmt.Sprintf("Index doesn't exist?")
		return nil, fmt.Errorf("%s", eMessage)
	}

	metrics := &InterfaceMetrics{
		Name:      string(result.Variables[0].Value.([]uint8)),
		In:        result.Variables[1].Value.(uint),
		Out:       result.Variables[2].Value.(uint),
		HCIn:      result.Variables[3].Value.(uint64),
		HCOut:     result.Variables[4].Value.(uint64),
		Speed:     result.Variables[5].Value.(uint),
		HighSpeed: result.Variables[6].Value.(uint),
		Latency:   latency,
		Timestamp: time.Now(),
	}

	return metrics, nil
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
	avgLatency := (first.Latency + second.Latency) / 2
	// Calc rates
	in := (second.In - first.In) / uint(period)
	out := (second.Out - first.Out) / uint(period)
	hcIn := (second.HCIn - first.HCIn) / uint64(period)
	hcOut := (second.HCOut - first.HCOut) / uint64(period)
	// Convert to scale
	intIn, intInUnit := convertToScale(uint64(in))
	intOut, intOutUnit := convertToScale(uint64(out))
	intHCIn, intHCInUnit := convertToScale(hcIn)
	intHCOut, intHCOutUnit := convertToScale(hcOut)
	// Craft message
	message := fmt.Sprintf("%s - In: %d %s Out: %d %s HCIn: %d %s HCOut: %d %s", intName, intIn, intInUnit, intOut, intOutUnit, intHCIn, intHCInUnit, intHCOut, intHCOutUnit)
	if enablePerf {
		checkResult.AddPerformanceData("snmp_latency", gomonitor.PerformanceMetric{Value: avgLatency.Seconds(), UnitOM: "s"})
		checkResult.AddPerformanceData("in", gomonitor.PerformanceMetric{Value: float64(in * 8), Warn: float64(warnIn), Crit: float64(critIn), Min: 0, Max: float64(first.Speed), UnitOM: "bps"})
		checkResult.AddPerformanceData("out", gomonitor.PerformanceMetric{Value: float64(out * 8), Warn: float64(warnOut), Crit: float64(critOut), Min: 0, Max: float64(first.Speed), UnitOM: "bps"})
		checkResult.AddPerformanceData("hc_in", gomonitor.PerformanceMetric{Value: float64(hcIn * 8), Warn: float64(warnIn), Crit: float64(critIn), Min: 0, Max: float64(first.Speed), UnitOM: "bps"})
		checkResult.AddPerformanceData("hc_out", gomonitor.PerformanceMetric{Value: float64(hcOut * 8), Warn: float64(warnOut), Crit: float64(critOut), Min: 0, Max: float64(first.Speed), UnitOM: "bps"})
	}

	if intIn > uint64(critIn) || intHCIn > uint64(critIn) {
		checkResult.SetResult(gomonitor.Critical, "Inbound exceeds threshold "+message)
	} else if intIn > uint64(warnIn) || intHCIn > uint64(warnIn) {
		checkResult.SetResult(gomonitor.Warning, "Inbound exceeds threshold "+message)
	} else if intOut > uint64(critOut) || intHCOut > uint64(critOut) {
		checkResult.SetResult(gomonitor.Critical, "Outbound exceeds threshold "+message)
	} else if intOut > uint64(warnOut) || intHCOut > uint64(warnOut) {
		checkResult.SetResult(gomonitor.Warning, "Outbound exceeds threshold "+message)
	} else {
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
	warnIn := flag.Int("warnIn", 0, "Warning level for inbound in bps. Default is 0.")
	critIn := flag.Int("critIn", 0, "Critical level for inbound in bps. Default is 0.")
	warnOut := flag.Int("warnOut", 0, "Warning level for outbound in bps. Default is 0.")
	critOut := flag.Int("critOut", 0, "Critical level for outbound bps. Default is 0.")
	flag.Parse()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
	}

	measure1, err1 := GetInterfaceMetrics(&snmpClient, *index)
	if err1 != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return data when measuring metrics. %s", snmpClient.Target, err1)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		checkResult.SendResult()
	}

	// delay
	time.Sleep(time.Duration(*delay) * time.Second)

	measure2, err2 := GetInterfaceMetrics(&snmpClient, *index)
	if err2 != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return data when measuring metrics. %s", snmpClient.Target, err2)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		checkResult.SendResult()
	}

	// Calculate current usage and determine thresholds
	result := DetermineInterfaceUsage(*measure1, *measure2, *warnIn, *warnOut, *critIn, *critOut, *enablePerfData)
	result.SendResult()
}
