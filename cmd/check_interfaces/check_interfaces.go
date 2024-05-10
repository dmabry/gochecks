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
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/dmabry/gochecks/internal"
	"github.com/dmabry/gomonitor"
	"regexp"
	"strconv"
	"strings"
)

type InterfaceDetail struct {
	Description    string
	Name           string
	Alias          string
	Type           int
	Index          int
	Speed          uint
	HighSpeed      uint
	OperStatus     int
	AdminStatus    int
	MTU            int
	InOctets       uint
	OutOctets      uint
	HCInOctets     uint64
	HCOutOctets    uint64
	HCInUcastPkts  uint64
	HCOutUcastPkts uint64
	InDiscards     uint
	OutDiscards    uint
	InErrors       uint
	OutErrors      uint
	InUcastPkts    uint
	OutUcastPkts   uint
	InNUcastPkts   uint
	OutNUcastPkts  uint
	LastChange     uint32
	PhysAddress    string
}

func CheckInterfaceMetrics(snmpClient *snmp.Client) *gomonitor.CheckResult {
	//baseOID := "1.3.6.1.2.1.2.2.1" // OID for IF-MIB::ifEntry (contains fields like ifDescr, ifType, ifSpeed, etc.)
	baseOIDs := []string{"1.3.6.1.2.1.2.2", "1.3.6.1.2.1.31.1.1.1"} // IF-MIB::ifEntry and ifXTable OIDs

	// Prepare data structure for holding interface details
	interfaces := make(map[int]*InterfaceDetail)

	checkResult := gomonitor.NewCheckResult()

	for _, baseOID := range baseOIDs {
		result, _, err := snmpClient.Walk(baseOID)
		if err != nil {
			eMessage := fmt.Sprintf("SNMP target %s failed to return data for requested OID: %w", snmpClient.Target, err)
			checkResult.SetResult(gomonitor.Critical, eMessage)
			return checkResult
		}
		for oid, value := range result {
			fields := strings.Split(oid, ".")
			// The index for each interface
			index, err2 := strconv.Atoi(fields[len(fields)-1])
			if err2 != nil {
				eMessage := fmt.Sprintf("failed to convert interface index to int: %w", err2)
				checkResult.SetResult(gomonitor.Critical, eMessage)
				return checkResult
			}

			// Remove the index from the OID
			oidWithoutIndex := strings.Join(fields[:len(fields)-1], ".")

			// Prepare each interface for holding details
			if _, ok := interfaces[index]; !ok {
				interfaces[index] = &InterfaceDetail{}
			}

			ifaceDetails := interfaces[index]

			// Match on the complete OID, excluding the index
			switch oidWithoutIndex {
			case ".1.3.6.1.2.1.2.2.1.2":
				if val, ok := value.([]byte); ok {
					ifaceDetails.Description = string(val)
				} else {
					fmt.Printf("ifDescr was not of type string: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.3":
				if val, ok := value.(int); ok {
					ifaceDetails.Type = val
				} else {
					fmt.Printf("ifType was not of type int: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.1":
				if val, ok := value.(int); ok {
					ifaceDetails.Index = val
				} else {
					fmt.Printf("ifIndex was not of type int: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.4":
				if val, ok := value.(int); ok {
					ifaceDetails.MTU = val
				} else {
					fmt.Printf("ifMTU was not of type int: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.5":
				if val, ok := value.(uint); ok {
					ifaceDetails.Speed = val
				} else {
					fmt.Printf("ifSpeed was not of type uint: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.13":
				if val, ok := value.(uint); ok {
					ifaceDetails.InDiscards = val
				} else {
					fmt.Printf("ifInDiscards was not of type uint: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.15":
				if val, ok := value.(uint); ok {
					ifaceDetails.OutUcastPkts = val
				} else {
					fmt.Printf("ifOutUcastPkts was not of type uint: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.19":
				if val, ok := value.(uint); ok {
					ifaceDetails.OutDiscards = val
				} else {
					fmt.Printf("ifOutDiscards was not of type uint: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.7":
				if val, ok := value.(int); ok {
					ifaceDetails.AdminStatus = val
				} else {
					fmt.Printf("ifAdminStatus was not of type int: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.8":
				if val, ok := value.(int); ok {
					ifaceDetails.OperStatus = val
				} else {
					fmt.Printf("ifOperStatus was not of type int: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.9":
				if val, ok := value.(uint32); ok {
					ifaceDetails.LastChange = val
				} else {
					fmt.Printf("ifLastChange was not of type uint32: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.10":
				if val, ok := value.(uint); ok {
					ifaceDetails.InOctets = val
				} else {
					fmt.Printf("ifInOctets was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.16":
				if val, ok := value.(uint); ok {
					ifaceDetails.OutOctets = val
				} else {
					fmt.Printf("ifOutOctets was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.14":
				if val, ok := value.(uint); ok {
					ifaceDetails.InErrors = val
				} else {
					fmt.Printf("ifInErrors was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.20":
				if val, ok := value.(uint); ok {
					ifaceDetails.OutErrors = val
				} else {
					fmt.Printf("ifOutErrors was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.11":
				if val, ok := value.(uint); ok {
					ifaceDetails.InUcastPkts = val
				} else {
					fmt.Printf("ifInUcastPkts was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.17":
				if val, ok := value.(uint); ok {
					ifaceDetails.OutUcastPkts = val
				} else {
					fmt.Printf("ifOutUcastPkts was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.2.2.1.6":
				if val, ok := value.([]byte); ok {
					ifaceDetails.PhysAddress = hex.EncodeToString(val)
				} else {
					fmt.Printf("ifPhysAddress was not of type []byte: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.1":
				if val, ok := value.([]byte); ok {
					ifaceDetails.Name = string(val)
				} else {
					fmt.Printf("ifName was not of type []byte: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.18":
				if val, ok := value.([]byte); ok {
					ifaceDetails.Alias = string(val)
				} else {
					fmt.Printf("ifAlias was not of type []byte: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.6":
				if val, ok := value.(uint64); ok {
					ifaceDetails.HCInOctets = val
				} else {
					fmt.Printf("HCInOctets was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.10":
				if val, ok := value.(uint64); ok {
					ifaceDetails.HCOutOctets = val
				} else {
					fmt.Printf("HCOutOctets was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.13":
				if val, ok := value.(uint64); ok {
					ifaceDetails.HCInUcastPkts = val
				} else {
					fmt.Printf("ifHCInUcastPkts was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.11":
				if val, ok := value.(uint64); ok {
					ifaceDetails.HCOutUcastPkts = val
				} else {
					fmt.Printf("ifHCOutUcastPkts was not of type uint64: %T -> %v\n", value, value)
				}
			case ".1.3.6.1.2.1.31.1.1.1.15":
				if val, ok := value.(uint); ok {
					ifaceDetails.HighSpeed = val
				} else {
					fmt.Printf("ifHighSpeed was not of type uint: %T -> %v\n", value, value)
				}
				/*default: //Should be things we don't care about.  TODO: I might make this part of a debug
				fmt.Printf("Unknown Type: OID: %s - %T -> %v\n", oid, value, value)*/
			}
		}
	}

	message := ""
	for index, iface := range interfaces {
		message += fmt.Sprintf("Interface index: %d\nDescription: %s\nAlias: %s\nName: %s\nType: %d\nSpeed: %d\nOperStatus: %d\nAdminStatus: %d\nInOctets: %d\nOutOctets: %d\nHCInOctets: %d\nHCOutOctets: %d\nHCInUcastPkts: %d\nHCOutUcastPkts: %d\nInErrors: %d\nOutErrors: %d\nInUcastPkts: %d\nOutUcastPkts: %d\nInNUcastPkts: %d\nOutNUcastPkts: %d\nLastChange: %d\nPhysAddress: %s\n\n",
			index,
			iface.Description,
			iface.Alias,
			iface.Name,
			iface.Type,
			iface.Speed,
			iface.OperStatus,
			iface.AdminStatus,
			iface.InOctets,
			iface.OutOctets,
			iface.HCInOctets,
			iface.HCOutOctets,
			iface.HCInUcastPkts,
			iface.HCOutUcastPkts,
			iface.InErrors,
			iface.OutErrors,
			iface.InUcastPkts,
			iface.OutUcastPkts,
			iface.InNUcastPkts,
			iface.OutNUcastPkts,
			iface.LastChange,
			iface.PhysAddress)
	}
	checkResult.SetResult(gomonitor.OK, message)
	return checkResult
}

func CheckSysDescr(snmpClient *snmp.Client, expectedSysDescrRegExp string, enablePerfData bool) *gomonitor.CheckResult {
	oids := []string{"1.3.6.1.2.1.1.1.0"}

	result, latency, err := snmpClient.GetValue(oids)
	if err != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return data for requested OID.", snmpClient.Target)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		return checkResult
	}

	checkResult := gomonitor.NewCheckResult()
	sysDescr := string(result.Variables[0].Value.([]uint8))

	// Compare result with expected sysDescr using regexp
	if expectedSysDescrRegExp != "" {
		match, err := regexp.MatchString(expectedSysDescrRegExp, sysDescr)
		if err != nil || !match {
			eMessage := fmt.Sprintf("sysDescr does not match expected pattern '%s'. Got: %s", expectedSysDescrRegExp, sysDescr)
			checkResult.SetResult(gomonitor.Critical, eMessage)
			return checkResult
		}
	}
	message := fmt.Sprintf("%s", sysDescr)
	checkResult.SetResult(gomonitor.OK, message)

	if enablePerfData {
		// If performance data is enabled, add SNMP latency to the check result
		checkResult.AddPerformanceData("latency", gomonitor.PerformanceMetric{Value: latency.Seconds(), UnitOM: "seconds"})
	}
	return checkResult
}

// main is the entry point of the program. It parses command-line flags, creates an SNMP client,
// and performs a check on the target SNMP device using the CheckSysDescr function.
// The result of the check is then sent using the SendResult method.
func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	// enablePerfData := flag.Bool("enablePerfData", false, "Enable performance data. Default is false.")
	flag.Parse()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
	}
	result := CheckInterfaceMetrics(&snmpClient)

	result.SendResult()
}
