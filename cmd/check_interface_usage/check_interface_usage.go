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
)

func convertToScale(value uint64) float64 {
	kbps := (float64(value * 8)) / 1000 // convert octets to Kbps
	if kbps < 1000 {
		return kbps
	}

	mbps := kbps / 1000 // convert Kbps to Mbps
	if mbps < 1000 {
		return mbps
	}

	gbps := mbps / 1000 // convert Mbps to Gbps
	return gbps
}

func CheckInterfaceUsage(snmpClient *snmp.Client, index int) *gomonitor.CheckResult {
	strIndex := strconv.Itoa(index)
	oidName := fmt.Sprintf("%s.%s", interfaces.OIDIfName, strIndex)
	oidHCIn := fmt.Sprintf("%s.%s", interfaces.OIDIfHCInOctets, strIndex)
	oidHCOut := fmt.Sprintf("%s.%s", interfaces.OIDIfHCOutOctets, strIndex)
	oidIn := fmt.Sprintf("%s.%s", interfaces.OIDIfInOctets, strIndex)
	oidOut := fmt.Sprintf("%s.%s", interfaces.OIDIfOutOctets, strIndex)
	usageOIDs := []string{oidName, oidIn, oidOut, oidHCIn, oidHCOut}

	// Prepare data structure for holding interface details

	checkResult := gomonitor.NewCheckResult()
	result, latency, err := snmpClient.GetValue(usageOIDs)
	if err != nil {
		eMessage := fmt.Sprintf("SNMP target %s failed to return data for requested OID: %w", snmpClient.Target, err)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		return checkResult
	}
	
	intName := string(result.Variables[0].Value.([]uint8))
	intIn := result.Variables[1].Value
	intOut := result.Variables[2].Value
	intHCIn := result.Variables[3].Value
	intHCOut := result.Variables[4].Value
	message := fmt.Sprintf("%s - In: %d Out: %d HCIn: %d HCOut: %d", intName, intIn, intOut, intHCIn, intHCOut)
	checkResult.AddPerformanceData("snmp_latency", gomonitor.PerformanceMetric{Value: latency.Seconds(), UnitOM: "seconds"})
	checkResult.SetResult(gomonitor.OK, message)
	return checkResult
}

// main is the entry point of the program. It parses command-line flags, creates an SNMP client,
// and performs a check on the target SNMP device using the CheckInterfaceMetrics function.
// The result of the check is then sent using the SendResult method.
func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	index := flag.Int("index", 1, "The index of the Interface")
	// enablePerfData := flag.Bool("enablePerfData", false, "Enable performance data. Default is false.")
	flag.Parse()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
	}
	result := CheckInterfaceUsage(&snmpClient, *index)

	result.SendResult()
}
