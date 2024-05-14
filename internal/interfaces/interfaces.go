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
