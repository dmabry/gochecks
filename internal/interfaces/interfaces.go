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

package interfaces

import (
	"encoding/json"
	"fmt"
)

// OID represents an SNMP Object Identifier as a string type.
type OID string

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

func (ifaceDetail *InterfaceDetail) ToString(index int) string {
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

func (ifaceDetail *InterfaceDetail) ToJsonString() (string, error) {
	jsonBytes, err := json.Marshal(ifaceDetail)
	if err != nil {
		return "", err
	}

	jsonString := string(jsonBytes)
	return jsonString, nil
}

// SetField sets a field on InterfaceDetail based on the given OID and value.
// It performs strict type checking and returns an error if the value type does
// not match the expected type for the OID or if the OID is unrecognized.
func (ifaceDetail *InterfaceDetail) SetField(oid OID, value interface{}) error {
	switch oid {
	case OIDIfDescr:
		if val, ok := value.(string); ok {
			ifaceDetail.Description = val
			return nil
		}
		return fmt.Errorf("OID %s requires string, got %T", oid, value)

	case OIDIfName:
		if val, ok := value.(string); ok {
			ifaceDetail.Name = val
			return nil
		}
		return fmt.Errorf("OID %s requires string, got %T", oid, value)

	case OIDIfAlias:
		if val, ok := value.(string); ok {
			ifaceDetail.Alias = val
			return nil
		}
		return fmt.Errorf("OID %s requires string, got %T", oid, value)

	case OIDIfPhysAddress:
		if val, ok := value.([]byte); ok {
			var parts []string
			for _, b := range val {
				parts = append(parts, fmt.Sprintf("%02x", b))
			}
			ifaceDetail.PhysAddress = fmt.Sprintf("%s:%s:%s:%s:%s:%s", parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
			return nil
		}
		return fmt.Errorf("OID %s requires []byte, got %T", oid, value)

	case OIDIfIndex:
		switch v := value.(type) {
		case int:
			ifaceDetail.Index = v
			return nil
		case int64:
			ifaceDetail.Index = int(v)
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfType:
		if val, ok := value.(int); ok {
			ifaceDetail.Type = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfMTU:
		if val, ok := value.(int); ok {
			ifaceDetail.MTU = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfSpeed:
		switch v := value.(type) {
		case uint:
			ifaceDetail.Speed = v
			return nil
		case int:
			if v < 0 {
				return fmt.Errorf("OID %s requires non-negative uint, got %d", oid, v)
			}
			ifaceDetail.Speed = uint(v)
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfHighSpeed:
		if val, ok := value.(uint); ok {
			ifaceDetail.HighSpeed = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOperStatus:
		if val, ok := value.(int); ok {
			ifaceDetail.OperStatus = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfAdminStatus:
		if val, ok := value.(int); ok {
			ifaceDetail.AdminStatus = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfInOctets:
		if val, ok := value.(uint); ok {
			ifaceDetail.InOctets = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutOctets:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutOctets = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfHCInOctets:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCInOctets = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfHCOutOctets:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCOutOctets = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfInUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.InUcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutUcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfHCInUcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCInUcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfHCOutUcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCOutUcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfInMulticastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.InMulticastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutMulticastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutMulticastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfHCInMulticastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCInMulticastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfHCOutMulticastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCOutMulticastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfInBroadcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.InBroadcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutBroadcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutBroadcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfHCInBroadcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCInBroadcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfHCOutBroadcastPkts:
		if val, ok := value.(uint64); ok {
			ifaceDetail.HCOutBroadcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint64, got %T", oid, value)

	case OIDIfInNUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.InNUcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutNUcastPkts:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutNUcastPkts = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfInErrors:
		if val, ok := value.(uint); ok {
			ifaceDetail.InErrors = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutErrors:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutErrors = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfInDiscards:
		if val, ok := value.(uint); ok {
			ifaceDetail.InDiscards = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfOutDiscards:
		if val, ok := value.(uint); ok {
			ifaceDetail.OutDiscards = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint, got %T", oid, value)

	case OIDIfLastChange:
		if val, ok := value.(uint32); ok {
			ifaceDetail.LastChange = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint32, got %T", oid, value)

	case OIDIfLinkUpDownTrapEnable:
		if val, ok := value.(int); ok {
			ifaceDetail.LinkUpDownTrapEnable = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfPromiscuousMode:
		if val, ok := value.(int); ok {
			ifaceDetail.PromiscuousMode = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfConnectorPresent:
		if val, ok := value.(int); ok {
			ifaceDetail.ConnectorPresent = val
			return nil
		}
		return fmt.Errorf("OID %s requires int, got %T", oid, value)

	case OIDIfCounterDiscontinuityTime:
		if val, ok := value.(uint32); ok {
			ifaceDetail.CounterDiscontinuityTime = val
			return nil
		}
		return fmt.Errorf("OID %s requires uint32, got %T", oid, value)

	default:
		return fmt.Errorf("unknown OID: %s", oid)
	}
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
	OIDIfHCInOctets               = ".1.3.6.1.2.1.31.1.1.1.6"
	OIDIfHCOutOctets              = ".1.3.6.1.2.1.31.1.1.1.10"
	OIDIfInUcastPkts              = ".1.3.6.1.2.1.2.2.1.11"
	OIDIfOutUcastPkts             = ".1.3.6.1.2.1.2.2.1.17"
	OIDIfHCInUcastPkts            = ".1.3.6.1.2.1.31.1.1.1.7"
	OIDIfHCOutUcastPkts           = ".1.3.6.1.2.1.31.1.1.1.11"
	OIDIfInBroadcastPkts          = ".1.3.6.1.2.1.31.1.1.1.3"
	OIDIfOutBroadcastPkts         = ".1.3.6.1.2.1.31.1.1.1.5"
	OIDIfHCInBroadcastPkts        = ".1.3.6.1.2.1.31.1.1.1.9"
	OIDIfHCOutBroadcastPkts       = ".1.3.6.1.2.1.31.1.1.1.13"
	OIDIfInMulticastPkts          = ".1.3.6.1.2.1.31.1.1.1.2"
	OIDIfOutMulticastPkts         = ".1.3.6.1.2.1.31.1.1.1.4"
	OIDIfHCInMulticastPkts        = ".1.3.6.1.2.1.31.1.1.1.8"
	OIDIfHCOutMulticastPkts       = ".1.3.6.1.2.1.31.1.1.1.12"
	OIDIfInNUcastPkts             = ".1.3.6.1.2.1.2.2.1.12"
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
