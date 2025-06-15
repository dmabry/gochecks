

/*
   Copyright 2024 David Mabry

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-,2.0

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
	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/dmabry/gomonitor"
	"strconv"
	"time"
)

// BgpPeer represents a BGP peer with admin and operational status
type BgpPeer struct {
	Index          int
	AdminStatus    int // 1=enabled, 2=disabled
	OperationalStatus int // 1=up, 2=down
}

// GetBgpPeers retrieves BGP peer information using SNMP from CISCO-BGP4-MIB
func GetBgpPeers(snmpClient *snmp.Client) ([]BgpPeer, error) {
	// OIDs from CISCO-BGP4-MIB.my
	adminStatusOID := ".1.3.6.1.2.1.15.4.1.8" // bgpPeerAdminStatus
	operationalStatusOID := ".1.3.6.1.2.1.15.4.1.9" // bgpPeerState
	
	// First get all peer indices
	indexOID := ".1.3.6.1.2.1.15.4.1.1" // bgpPeerIdentifier
	indexResult, _, err := snmpClient.GetTable(indexOID)
	if err != nil {
		return nil, fmt.Errorf("failed to get BGP peer indices: %w", err)
	}
	
	var peers []BgpPeer
	
	for _, index := range indexResult.Variables {
		indexStr := strconv.Itoa(int(index.Value.(int64)))
		
		adminOID := fmt.Sprintf("%s.%s", adminStatusOID, indexStr)
		operationalOID := fmt.Sprintf("%s.%s", operationalStatusOID, indexStr)
		
		result, _, err := snmpClient.GetValues([]string{adminOID, operationalOID})
		if err != nil {
			continue // Skip peers that can't be queried
		}
		
		adminStatus := int(result.Variables[0].Value.(int64))
		operationalStatus := int(result.Variables[1].Value.(int64))
		
		peers = append(peers, BgpPeer{
			Index:          int(index.Value.(int64)),
			AdminStatus:    adminStatus,
			OperationalStatus: operationalStatus,
		})
	}
	
	return peers, nil
}

func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	flag.Parse()
	
	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
	}
	
	peers, err := GetBgpPeers(&snmpClient)
	if err != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return BGP peer data: %s", snmpClient.Target, err)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		checkResult.SendResult()
		return
	}
	
	mismatchCount := 0
	for _, peer := range peers {
		if peer.AdminStatus != peer.OperationalStatus {
			mismatchCount++
		}
	}
	
	if mismatchCount > 0 {
		checkResult := gomonitor.NewCheckResult()
		message := fmt.Sprintf("Found %d BGP peer(s) with admin status mismatch", mismatchCount)
		checkResult.SetResult(gomonitor.Critical, message)
		checkResult.AddPerformanceData("mismatched_peers", gomonitor.PerformanceMetric{Value: float64(mismatchCount), UnitOM: "count"})
		checkResult.SendResult()
	} else {
		checkResult := gomonitor.NewCheckResult()
		message := fmt.Sprintf("All %d BGP peers have matching admin and operational status", len(peers))
		checkResult.SetResult(gomonitor.OK, message)
		checkResult.AddPerformanceData("total_peers", gomonitor.PerformanceMetric{Value: float64(len(peers)), UnitOM: "count"})
		checkResult.SendResult()
	}
}

