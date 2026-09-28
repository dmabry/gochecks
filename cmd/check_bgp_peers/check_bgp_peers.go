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
	"strings"

	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/dmabry/gomonitor"
	"github.com/gosnmp/gosnmp"
)

// BGP4-MIB OIDs (RFC 4273). The peer table is indexed by the remote peer
// address, so the admin status and state columns are queried per peer by
// appending the row index from the identifier column walk.
const (
	oidBgpPeerIdentifier = ".1.3.6.1.2.1.15.4.1.1"
	oidBgpPeerAdminState = ".1.3.6.1.2.1.15.4.1.8"
	oidBgpPeerState      = ".1.3.6.1.2.1.15.4.1.9"
)

// bgpPeerAdminStatus enum values.
const (
	bgpAdminStop  = 1
	bgpAdminStart = 2
)

// bgpPeerState enum values.
const (
	bgpStateIdle        = 1
	bgpStateConnect     = 2
	bgpStateActive      = 3
	bgpStateOpensent    = 4
	bgpStateOpenconfirm = 5
	bgpStateEstablished = 6
)

// BgpPeer represents a BGP peer with its identifier, administrative status,
// and session state.
type BgpPeer struct {
	// Identifier is the remote peer address (bgpPeerIdentifier).
	Identifier string
	// AdminStatus is bgpPeerAdminStatus: 1=stop, 2=start.
	AdminStatus int
	// State is bgpPeerState: 1=idle, 2=connect, 3=active, 4=opensent,
	// 5=openconfirm, 6=established.
	State int
}

// hasAdminStateMismatch reports whether a peer's administrative status and
// session state are inconsistent: administratively enabled but the session is
// not established, or administratively stopped while the session is up.
func (bp BgpPeer) hasAdminStateMismatch() bool {
	switch {
	case bp.AdminStatus == bgpAdminStart && bp.State != bgpStateEstablished:
		return true
	case bp.AdminStatus == bgpAdminStop && bp.State == bgpStateEstablished:
		return true
	default:
		return false
	}
}

// CheckBgpPeers checks the BGP peer table of an SNMP target and reports peers
// whose administrative status and session state are inconsistent.
//
// The function walks the bgpPeerIdentifier column to enumerate peers, then
// queries bgpPeerAdminStatus and bgpPeerState for each peer. Peers that cannot
// be queried are skipped. If an error occurs during the walk, a Critical check
// result is returned with the error message.
//
// If any peer has an admin status mismatch, a Critical result is returned
// naming the count of mismatched peers. Otherwise an OK result is returned
// with the total peer count.
//
// If enablePerfData is true, the total peer count and mismatched peer count
// are added to the performance data of the check result.
//
// Example usage:
//
//	snmpClient := snmp.Client{
//	    Target:    "127.0.0.1",
//	    Community: "public",
//	}
//	result := CheckBgpPeers(&snmpClient, true)
//	result.SendResult()
func CheckBgpPeers(snmpClient *snmp.Client, enablePerfData bool) *gomonitor.CheckResult {
	peers, err := GetBgpPeers(snmpClient)
	if err != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return BGP peer data: %s", snmpClient.Target, err)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		return checkResult
	}
	return EvaluateBgpPeers(peers, enablePerfData)
}

// EvaluateBgpPeers builds the check result from a set of BGP peers. If any
// peer has an admin status mismatch, a Critical result is returned naming the
// count of mismatched peers. Otherwise an OK result is returned with the total
// peer count. If enablePerfData is true, the total peer count and mismatched
// peer count are added to the performance data.
func EvaluateBgpPeers(peers []BgpPeer, enablePerfData bool) *gomonitor.CheckResult {
	mismatchCount := 0
	for _, peer := range peers {
		if peer.hasAdminStateMismatch() {
			mismatchCount++
		}
	}

	checkResult := gomonitor.NewCheckResult()
	if mismatchCount > 0 {
		message := fmt.Sprintf("Found %d BGP peer(s) with admin status mismatch", mismatchCount)
		checkResult.SetResult(gomonitor.Critical, message)
	} else {
		message := fmt.Sprintf("All %d BGP peers have matching admin and operational status", len(peers))
		checkResult.SetResult(gomonitor.OK, message)
	}
	if enablePerfData {
		checkResult.AddPerformanceData("total_peers", gomonitor.PerformanceMetric{Value: float64(len(peers)), UnitOM: "count"})
		checkResult.AddPerformanceData("mismatched_peers", gomonitor.PerformanceMetric{Value: float64(mismatchCount), UnitOM: "count"})
	}
	return checkResult
}

// GetBgpPeers retrieves BGP peer information using SNMP from BGP4-MIB. It
// walks the bgpPeerIdentifier column to enumerate peers and queries the admin
// status and session state columns for each peer. Peers that cannot be queried
// are skipped.
func GetBgpPeers(snmpClient *snmp.Client) ([]BgpPeer, error) {
	identifierColumn, _, err := snmpClient.Walk(context.TODO(), oidBgpPeerIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get BGP peer indices: %w", err)
	}

	var peers []BgpPeer
	for oid, value := range identifierColumn {
		// IPAddress PDUs decode to a dotted-quad string; skip non-address
		// values such as the nil value buggy devices return.
		identifier, ok := value.(string)
		if !ok {
			continue
		}

		// The row index is the remote peer address bytes appended to the
		// identifier column OID; reuse it to address the status columns.
		rowIndex := strings.TrimPrefix(oid, oidBgpPeerIdentifier)
		statusOIDs := []string{oidBgpPeerAdminState + rowIndex, oidBgpPeerState + rowIndex}

		result, _, err := snmpClient.GetValue(context.TODO(), statusOIDs)
		if err != nil || len(result.Variables) < 2 {
			continue // Skip peers that can't be queried
		}
		adminStatus, ok := result.Variables[0].Value.(int)
		if !ok {
			continue // Skip peers with unexpected admin status data
		}
		state, ok := result.Variables[1].Value.(int)
		if !ok {
			continue // Skip peers with unexpected state data
		}

		peers = append(peers, BgpPeer{
			Identifier:  identifier,
			AdminStatus: adminStatus,
			State:       state,
		})
	}
	return peers, nil
}

// main is the entry point of the program. It parses command-line flags, creates an SNMP client,
// and performs a check on the target SNMP device using the CheckBgpPeers function.
// The result of the check is then sent using the SendResult method.
func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	enablePerfData := flag.Bool("enablePerfData", false, "Enable performance data. Default is false.")
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
	result := CheckBgpPeers(&snmpClient, *enablePerfData)
	result.SendResult()
}
