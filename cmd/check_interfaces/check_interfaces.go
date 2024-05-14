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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dmabry/gochecks/internal"
	"github.com/dmabry/gomonitor"
	"log"
	"strconv"
	"strings"
)

type InterfaceDetail struct {
	// Basic Info
	Description string
	Name        string
	Alias       string
	PhysAddress string

	// Identification and Types
	Index int
	Type  int
	MTU   int

	// Speeds
	Speed     uint
	HighSpeed uint

	// Status
	OperStatus  int
	AdminStatus int

	// Octets
	InOctets    uint
	OutOctets   uint
	HCInOctets  uint64
	HCOutOctets uint64

	// Packets
	InUcastPkts        uint
	OutUcastPkts       uint
	HCInUcastPkts      uint64
	HCOutUcastPkts     uint64
	InMulticastPkts    uint
	OutMulticastPkts   uint
	HCInMulticastPkts  uint64
	HCOutMulticastPkts uint64
	InBroadcastPkts    uint
	OutBroadcastPkts   uint
	HCInBroadcastPkts  uint64
	HCOutBroadcastPkts uint64
	InNUcastPkts       uint
	OutNUcastPkts      uint

	// Errors and Discards
	InErrors    uint
	OutErrors   uint
	InDiscards  uint
	OutDiscards uint

	// Miscellaneous
	LastChange               uint32
	LinkUpDownTrapEnable     int
	PromiscuousMode          int
	ConnectorPresent         int
	CounterDiscontinuityTime uint32
}

func (ifaceDetail *InterfaceDetail) renderToString(index int) string {
	const (
		outputFormat = "Interface index: %d\nDescription: %s\nAlias: %s\nName: %s\nType: %d\nSpeed: %d\nHighSpeed: %d\nOperStatus: %d\nAdminStatus: %d\nInOctets: %d\nOutOctets: %d\nHCInOctets: %d\nHCOutOctets: %d\nHCInUcastPkts: %d\nHCOutUcastPkts: %d\nInErrors: %d\nOutErrors: %d\nInUcastPkts: %d\nOutUcastPkts: %d\nInNUcastPkts: %d\nOutNUcastPkts: %d\nPromiscuousMode: %d\nLastChange: %d\nPhysAddress: %s\n\n"
	)
	return fmt.Sprintf(outputFormat,
		index,
		ifaceDetail.Description,
		ifaceDetail.Alias,
		ifaceDetail.Name,
		ifaceDetail.Type,
		ifaceDetail.Speed,
		ifaceDetail.HighSpeed,
		ifaceDetail.OperStatus,
		ifaceDetail.AdminStatus,
		ifaceDetail.InOctets,
		ifaceDetail.OutOctets,
		ifaceDetail.HCInOctets,
		ifaceDetail.HCOutOctets,
		ifaceDetail.HCInUcastPkts,
		ifaceDetail.HCOutUcastPkts,
		ifaceDetail.InErrors,
		ifaceDetail.OutErrors,
		ifaceDetail.InUcastPkts,
		ifaceDetail.OutUcastPkts,
		ifaceDetail.InNUcastPkts,
		ifaceDetail.OutNUcastPkts,
		ifaceDetail.PromiscuousMode,
		ifaceDetail.LastChange,
		ifaceDetail.PhysAddress)
}

func (ifaceDetail *InterfaceDetail) toJsonString() (string, error) {
	jsonBytes, err := json.Marshal(ifaceDetail)
	if err != nil {
		return "", err
	}

	jsonString := string(jsonBytes)
	return jsonString, nil
}

const (
	OIDIfDescr                    = ".1.3.6.1.2.1.2.2.1.2"
	OIDIfName                     = ".1.3.6.1.2.1.31.1.1.1.1"
	OIDIfAlias                    = ".1.3.6.1.2.1.31.1.1.1.18"
	OIDIfPhysAddress              = ".1.3.6.1.2.1.2.2.1.6"
	OIDIfIndex                    = ".1.3.6.1.2.1.2.2.1.1"
	OIDIfType                     = ".1.3.6.1.2.1.2.2.1.3"
	OIDIfMTU                      = ".1.3.6.1.2.1.2.2.1.4"
	OIDIfSpeed                    = ".1.3.6.1.2.1.2.2.1.5"
	OIDIfHighSpeed                = ".1.3.6.1.2.1.31.1.1.1.15"
	OIDIfOperStatus               = ".1.3.6.1.2.1.2.2.1.8"
	OIDIfAdminStatus              = ".1.3.6.1.2.1.2.2.1.7"
	OIDIfInOctets                 = ".1.3.6.1.2.1.2.2.1.10"
	OIDIfOutOctets                = ".1.3.6.1.2.1.2.2.1.16"
	OIDHCInOctets                 = ".1.3.6.1.2.1.31.1.1.1.6"
	OIDHCOutOctets                = ".1.3.6.1.2.1.31.1.1.1.10"
	OIDIfInUcastPkts              = ".1.3.6.1.2.1.2.2.1.11"
	OIDIfOutUcastPkts             = ".1.3.6.1.2.1.2.2.1.17"
	OIDHCInUcastPkts              = ".1.3.6.1.2.1.31.1.1.1.7"
	OIDHCOutUcastPkts             = ".1.3.6.1.2.1.31.1.1.1.11"
	OIDIfInBroadcastPkts          = ".1.3.6.1.2.1.31.1.1.1.3"
	OIDIfOutBroadcastPkts         = ".1.3.6.1.2.1.31.1.1.1.5"
	OIDIfHCInBroadcastPkts        = ".1.3.6.1.2.1.31.1.1.1.9"
	OIDIfHCOutBroadcastPkts       = ".1.3.6.1.2.1.31.1.1.1.13"
	OIDIfInMulticastPkts          = ".1.3.6.1.2.1.31.1.1.1.2"
	OIDIfOutMulticastPkts         = ".1.3.6.1.2.1.31.1.1.1.4"
	OIDIfHCInMulticastPkts        = ".1.3.6.1.2.1.31.1.1.1.8"
	OIDIfHCOutMulticastPkts       = ".1.3.6.1.2.1.31.1.1.1.12"
	OIDIfOutNUcastPkts            = ".1.3.6.1.2.1.2.2.1.15"
	OIDIfInErrors                 = ".1.3.6.1.2.1.2.2.1.14"
	OIDIfOutErrors                = ".1.3.6.1.2.1.2.2.1.20"
	OIDIfInDiscards               = ".1.3.6.1.2.1.2.2.1.13"
	OIDIfOutDiscards              = ".1.3.6.1.2.1.2.2.1.19"
	OIDIfLastChange               = ".1.3.6.1.2.1.2.2.1.9"
	OIDIfLinkUpDownTrapEnable     = ".1.3.6.1.2.1.31.1.1.1.14"
	OIDIfPromiscuousMode          = ".1.3.6.1.2.1.31.1.1.1.16"
	OIDIfConnectorPresent         = ".1.3.6.1.2.1.31.1.1.1.17"
	OIDIfCounterDiscontinuityTime = ".1.3.6.1.2.1.31.1.1.1.19"
)

func updateInterfaceDetails(ifaceDetails *InterfaceDetail, oid string, value interface{}) {
	switch oid {
	case OIDIfIndex:
		if val, ok := value.(int); ok {
			ifaceDetails.Index = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfDescr:
		if val, ok := value.([]byte); ok {
			ifaceDetails.Description = string(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case OIDIfType:
		if val, ok := value.(int); ok {
			ifaceDetails.Type = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfMTU:
		if val, ok := value.(int); ok {
			ifaceDetails.MTU = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfSpeed:
		if val, ok := value.(uint); ok {
			ifaceDetails.Speed = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfHighSpeed:
		if val, ok := value.(uint); ok {
			ifaceDetails.HighSpeed = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfPhysAddress:
		if val, ok := value.([]byte); ok {
			ifaceDetails.PhysAddress = hex.EncodeToString(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case OIDIfAdminStatus:
		if val, ok := value.(int); ok {
			ifaceDetails.AdminStatus = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfOperStatus:
		if val, ok := value.(int); ok {
			ifaceDetails.OperStatus = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfLastChange:
		if val, ok := value.(uint32); ok {
			ifaceDetails.LastChange = val
		} else {
			log.Printf("Value for OID %s is not of type uint32: %T -> %v\n", oid, value, value)
		}
	case OIDIfInOctets:
		if val, ok := value.(uint); ok {
			ifaceDetails.InOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfInUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.InUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfInDiscards:
		if val, ok := value.(uint); ok {
			ifaceDetails.InDiscards = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfInErrors:
		if val, ok := value.(uint); ok {
			ifaceDetails.InErrors = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutOctets:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutDiscards:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutDiscards = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutErrors:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutErrors = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutNUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutNUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfName:
		if val, ok := value.([]byte); ok {
			ifaceDetails.Name = string(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case OIDIfAlias:
		if val, ok := value.([]byte); ok {
			ifaceDetails.Alias = string(val)
		} else {
			log.Printf("Value for OID %s is not of type []byte: %T -> %v\n", oid, value, value)
		}
	case OIDHCInOctets:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDHCOutOctets:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutOctets = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDHCInUcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDHCOutUcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutUcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDIfHCInMulticastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDIfHCInBroadcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCInBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDIfHCOutBroadcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDIfLinkUpDownTrapEnable:
		if val, ok := value.(int); ok {
			ifaceDetails.LinkUpDownTrapEnable = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfConnectorPresent:
		if val, ok := value.(int); ok {
			ifaceDetails.ConnectorPresent = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutBroadcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfInBroadcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutBroadcastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfCounterDiscontinuityTime:
		if val, ok := value.(uint32); ok {
			ifaceDetails.CounterDiscontinuityTime = val
		} else {
			log.Printf("Value for OID %s is not of type uint32: %T -> %v\n", oid, value, value)
		}
	case OIDIfHCOutMulticastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetails.HCOutMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint64: %T -> %v\n", oid, value, value)
		}
	case OIDIfInMulticastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.InMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfOutMulticastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetails.OutMulticastPkts = val
		} else {
			log.Printf("Value for OID %s is not of type uint: %T -> %v\n", oid, value, value)
		}
	case OIDIfPromiscuousMode:
		if val, ok := value.(int); ok {
			ifaceDetails.PromiscuousMode = val
		} else {
			log.Printf("Value for OID %s is not of type int: %T -> %v\n", oid, value, value)
		}
	default:
		log.Printf("Unknown Type: OID: %s - %T -> %v\n", oid, value, value)
	}
}

func buildInterfaceDetailsMessage(interfaces map[int]*InterfaceDetail) string {
	var message strings.Builder
	for index, iface := range interfaces {
		message.WriteString(iface.renderToString(index))
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
			updateInterfaceDetails(ifaceDetails, oidWithoutIndex, value)
		}
	}

	message := buildInterfaceDetailsMessage(interfaces)
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
