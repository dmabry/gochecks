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
	"log"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

// buildInterfaceDetailsMessage builds a message string containing the interface details
// for each interface in the map of InterfaceDetail structures.
//
// Parameters:
//   - interfaces: A map representing the interface details, where the key is the index of the interface
//     and the value is a pointer to an InterfaceDetail structure.
//
// Returns:
//   - message: A string representation of the interface details, where each interface is represented
//     with its index and its corresponding InterfaceDetail structure converted to a string.
//
// Usage Example:
//
//	deviceInterfaces := make(map[int]*interfaces.InterfaceDetail)
//	// Fill the deviceInterfaces map with interface details...
//	message := buildInterfaceDetailsMessage(deviceInterfaces)
func buildInterfaceDetailsMessage(interfaces map[int]*interfaces.InterfaceDetail) string {
	var message strings.Builder
	for index, iface := range interfaces {
		message.WriteString(iface.ToString(index))
	}
	return message.String()
}

// CheckInterfaceMetrics retrieves interface details from the target SNMP device using the provided SNMP client.
// It walks the IF-MIB::ifEntry and ifXTable OIDs to gather information about each interface.
// The function populates an InterfaceDetail structure for each interface encountered and builds a message
// with the interface details. If any error occurs during the SNMP request, it will set the result to Critical
// and return the error message along with the check result. Otherwise, it sets the result to OK and returns
// the interface details message along with the check result.
func CheckInterfaceMetrics(snmpClient *snmp.Client) *gomonitor.CheckResult {
	baseOIDs := []string{"1.3.6.1.2.1.2.2", "1.3.6.1.2.1.31.1.1.1"} // IF-MIB::ifEntry and ifXTable OIDs

	// Prepare data structure for holding interface details
	deviceInterfaces := make(map[int]*interfaces.InterfaceDetail)

	checkResult := gomonitor.NewCheckResult()

	for _, baseOID := range baseOIDs {
		result, _, err := snmpClient.Walk(context.TODO(), baseOID)
		if err != nil {
			eMessage := fmt.Sprintf("SNMP target %s failed to return data for requested OID: %v", snmpClient.Target, err)
			checkResult.SetResult(gomonitor.Critical, eMessage)
			return checkResult
		}
		for oid, value := range result {
			fields := strings.Split(oid, ".")
			// The index for each interface
			index, err2 := strconv.Atoi(fields[len(fields)-1])
			if err2 != nil {
				eMessage := fmt.Sprintf("failed to convert interface index to int: %v", err2)
				checkResult.SetResult(gomonitor.Critical, eMessage)
				return checkResult
			}

			// Remove the index from the OID
			oidWithoutIndex := strings.Join(fields[:len(fields)-1], ".")

			// Prepare each interface for holding details
			if _, ok := deviceInterfaces[index]; !ok {
				deviceInterfaces[index] = &interfaces.InterfaceDetail{}
			}

			ifaceDetails := deviceInterfaces[index]
			// Match on the complete OID, excluding the index; log values
			// whose type does not fit the OID instead of failing the check.
			if err := ifaceDetails.SetField(interfaces.OID(oidWithoutIndex), value); err != nil {
				log.Printf("Value for OID %s: %v\n", oidWithoutIndex, err)
			}
		}
	}

	message := buildInterfaceDetailsMessage(deviceInterfaces)
	checkResult.SetResult(gomonitor.OK, message)
	return checkResult
}

// countInterfaces counts the interfaces in a rendered interface-details
// message (each interface block is separated by a blank line).
func countInterfaces(message string) int {
	lines := strings.Split(strings.TrimSpace(message), "\n\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// filterInterfacesByDescription filters the interface results by a specific interface description.
// If no interfaces match the description, it returns a Critical result. Otherwise, it returns
// an OK result with only the matching interfaces.
func filterInterfacesByDescription(result *gomonitor.CheckResult, ifaceDesc string) *gomonitor.CheckResult {
	// Create a new check result for filtered interfaces
	filteredResult := gomonitor.NewCheckResult()

	// Check if the original result is OK and has interface data
	if result.ExitCode == gomonitor.OK {
		message := result.Message
		lines := strings.Split(message, "\n\n")

		var filteredInterfaces []string

		for _, line := range lines {
			if strings.Contains(line, "Description: "+ifaceDesc) {
				filteredInterfaces = append(filteredInterfaces, line)
			}
		}

		if len(filteredInterfaces) > 0 {
			filteredMessage := strings.Join(filteredInterfaces, "\n\n")
			filteredResult.SetResult(gomonitor.OK, filteredMessage)
		} else {
			filteredResult.SetResult(gomonitor.Critical, fmt.Sprintf("No interfaces found with description: %s", ifaceDesc))
		}
	} else {
		// If original result is not OK, return it as-is
		filteredResult.SetResult(result.ExitCode, result.Message)
	}

	return filteredResult
}

// main is the entry point of the program. It parses command-line flags, creates an SNMP client,
// and performs a check on the target SNMP device using the CheckInterfaceMetrics function.
// The result of the check is then sent using the SendResult method.
func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	ifaceDesc := flag.String("iface", "", "Filter interfaces by description (optional).")
	enablePerfData := flag.Bool("enablePerfData", false, "Enable performance data. Adds the interface count as a performance metric.")
	snmpVersion := &snmp.SNMPVersionFlag{}
	flag.Var(snmpVersion, "snmpVersion", "SNMP protocol version: 2c or 3 (default 2c).")
	v3Username, v3AuthProtocol, v3AuthPassphrase, v3PrivProtocol, v3PrivPassphrase := snmp.AddV3Flags(flag.CommandLine)
	flag.Parse()

	version := snmpVersion.VersionOrDefault()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
		Version:   version,
	}

	if version == gosnmp.Version3 {
		if err := snmp.ApplyV3Flags(&snmpClient, *v3Username, v3AuthProtocol.Protocol, *v3AuthPassphrase, v3PrivProtocol.Protocol, *v3PrivPassphrase); err != nil {
			checkResult := gomonitor.NewCheckResult()
			checkResult.SetResult(gomonitor.Unknown, err.Error())
			checkResult.SendResult()
			return
		}
	}
	result := CheckInterfaceMetrics(&snmpClient)

	if *enablePerfData {
		ifaceCount := countInterfaces(result.Message)
		result.AddPerformanceData("interfaces", gomonitor.PerformanceMetric{Value: float64(ifaceCount)})
	}

	if *ifaceDesc != "" {
		filteredResult := filterInterfacesByDescription(result, *ifaceDesc)
		filteredResult.SendResult()
	} else {
		result.SendResult()
	}
}
