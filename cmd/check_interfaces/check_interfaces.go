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
	"github.com/dmabry/gochecks/internal/interfaces"
	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/dmabry/gomonitor"
	"log"
	"strconv"
	"strings"
)

// updateInterfaceDetails updates the corresponding field in ifaceDetails based on the provided OID and value.
// It logs an error message if the value is not of the expected type for the specified OID.
// Supported OIDs and their expected value types:
// - interfaces.OIDIfIndex: int
// - interfaces.OIDIfDescr: []byte
// - interfaces.OIDIfType: int
// - interfaces.OIDIfMTU: int
// - interfaces.OIDIfSpeed: uint
// - interfaces.OIDIfHighSpeed: uint
// - interfaces.OIDIfPhysAddress: []byte
// - interfaces.OIDIfAdminStatus: int
// - interfaces.OIDIfOperStatus: int
// - interfaces.OIDIfLastChange: uint32
// - interfaces.OIDIfInOctets: uint
// - interfaces.OIDIfInUcastPkts: uint
// - interfaces.OIDIfInDiscards: uint
// - interfaces.OIDIfInErrors: uint
// - interfaces.OIDIfOutOctets: uint
// - interfaces.OIDIfOutUcastPkts: uint
// - interfaces.OIDIfOutDiscards: uint
// - interfaces.OIDIfOutErrors: uint
// - interfaces.OIDIfOutNUcastPkts: uint
// - interfaces.OIDIfName: []byte
// - interfaces.OIDIfAlias: []byte
// - interfaces.OIDIfHCInOctets: uint64
// - interfaces.OIDIfHCOutOctets: uint64
// - interfaces.OIDIfHCInUcastPkts: uint64
// - interfaces.OIDIfHCOutUcastPkts: uint64
// - interfaces.OIDIfHCInMulticastPkts: uint64
// - interfaces.OIDIfHCInBroadcastPkts: uint64
func updateInterfaceDetails(ifaceDetails *interfaces.InterfaceDetail, oid string, value interface{}) {
	switch oid {
	case interfaces.OIDIfIndex:
		if val, ok := value.(int); ok {
			ifaceDetails.Index = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfDescr:
		if val, ok := value.([]byte); ok {
			ifaceDetails.Description = string(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfType:
		if val, ok := value.(int); ok {
			ifaceDetails.Type = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfMTU:
		if val, ok := value.(int); ok {
			ifaceDetails.MTU = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfSpeed:
		if val, ok := value.(uint); ok {
			ifaceDetails.Speed = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHighSpeed:
		if val, ok := value.(uint); ok {
			ifaceDetails.HighSpeed = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfPhysAddress:
		if val, ok := value.([]byte); ok {
			ifaceDetails.PhysAddress = hex.EncodeToString(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfAdminStatus:
		if val, ok := value.(int); ok {
			ifaceDetails.AdminStatus = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOperStatus:
		if val, ok := value.(int); ok {
			ifaceDetails.OperStatus = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfLastChange:
		if val, ok := value.(uint32); ok {
			ifaceDetails.LastChange = val
		} else {
			log.Printf("Value for OID %s is not of type uint32: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfInOctets:
		if val, ok := value.(uint); ok {
			ifaceDetails.InOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfInUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.InUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfInDiscards:
		if val, ok := value.(uint); ok {
			ifaceDetails.InDiscards = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfInErrors:
		if val, ok := value.(uint); ok {
			ifaceDetails.InErrors = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutOctets:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutDiscards:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutDiscards = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutErrors:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutErrors = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutNUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutNUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfName:
		if val, ok := value.([]byte); ok {
			ifaceDetails.Name = string(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfAlias:
		if val, ok := value.([]byte); ok {
			ifaceDetails.Alias = string(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCInOctets:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCOutOctets:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCInUcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCOutUcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCInMulticastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCInBroadcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCOutBroadcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfLinkUpDownTrapEnable:
		if val, ok := value.(int); ok {
			ifaceDetails.LinkUpDownTrapEnable = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfConnectorPresent:
		if val, ok := value.(int); ok {
			ifaceDetails.ConnectorPresent = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutBroadcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfInBroadcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfCounterDiscontinuityTime:
		if val, ok := value.(uint32); ok {
			ifaceDetails.CounterDiscontinuityTime = val
		} else {
			log.Printf("Value for OID %s is not of type uint32: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfHCOutMulticastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfInMulticastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.InMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfOutMulticastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case interfaces.OIDIfPromiscuousMode:
		if val, ok := value.(int); ok {
			ifaceDetails.PromiscuousMode = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	default:
		log.Printf("Unknown Type: OID: %s - %T -> %v\n", oid, value, value)
	}
}

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
			if _, ok := deviceInterfaces[index]; !ok {
				deviceInterfaces[index] = &interfaces.InterfaceDetail{}
			}

			ifaceDetails := deviceInterfaces[index]
			// Match on the complete OID, excluding the index
			updateInterfaceDetails(ifaceDetails, oidWithoutIndex, value)
		}
	}

	message := buildInterfaceDetailsMessage(deviceInterfaces)
	checkResult.SetResult(gomonitor.OK, message)
	return checkResult
}

// main is the entry point of the program. It parses command-line flags, creates an SNMP client,
// and performs a check on the target SNMP device using the CheckInterfaceMetrics function.
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
