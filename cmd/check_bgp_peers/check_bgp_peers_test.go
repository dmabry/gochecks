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
	"strings"
	"testing"

	"github.com/dmabry/gomonitor"
)

// TestHasAdminStateMismatch verifies the admin status / session state
// inconsistency rule against the BGP4-MIB enum values: administratively
// enabled peers (start) must be established, and an established session under
// an administratively stopped peer is a mismatch.
func TestHasAdminStateMismatch(t *testing.T) {
	tests := []struct {
		name       string
		adminState BgpPeer
		want       bool
	}{
		{name: "start and established", adminState: BgpPeer{AdminStatus: bgpAdminStart, State: bgpStateEstablished}, want: false},
		{name: "start and idle", adminState: BgpPeer{AdminStatus: bgpAdminStart, State: bgpStateIdle}, want: true},
		{name: "start and connect", adminState: BgpPeer{AdminStatus: bgpAdminStart, State: bgpStateConnect}, want: true},
		{name: "start and active", adminState: BgpPeer{AdminStatus: bgpAdminStart, State: bgpStateActive}, want: true},
		{name: "start and opensent", adminState: BgpPeer{AdminStatus: bgpAdminStart, State: bgpStateOpensent}, want: true},
		{name: "start and openconfirm", adminState: BgpPeer{AdminStatus: bgpAdminStart, State: bgpStateOpenconfirm}, want: true},
		{name: "stop and idle", adminState: BgpPeer{AdminStatus: bgpAdminStop, State: bgpStateIdle}, want: false},
		{name: "stop and established", adminState: BgpPeer{AdminStatus: bgpAdminStop, State: bgpStateEstablished}, want: true},
		{name: "unset admin status", adminState: BgpPeer{AdminStatus: 0, State: bgpStateEstablished}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.adminState.hasAdminStateMismatch(); got != tt.want {
				t.Errorf("hasAdminStateMismatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEvaluateBgpPeers verifies the check result for mixed peer sets: any
// mismatched peer produces Critical naming the count, consistent peers
// produce OK naming the total, and an empty peer table is OK.
func TestEvaluateBgpPeers(t *testing.T) {
	tests := []struct {
		name     string
		peers    []BgpPeer
		wantCode gomonitor.ExitCode
	}{
		{
			name: "all peers established",
			peers: []BgpPeer{
				{Identifier: "10.0.0.1", AdminStatus: bgpAdminStart, State: bgpStateEstablished},
				{Identifier: "10.0.0.2", AdminStatus: bgpAdminStart, State: bgpStateEstablished},
			},
			wantCode: gomonitor.OK,
		},
		{
			name: "session down under enabled peer",
			peers: []BgpPeer{
				{Identifier: "10.0.0.1", AdminStatus: bgpAdminStart, State: bgpStateEstablished},
				{Identifier: "10.0.0.2", AdminStatus: bgpAdminStart, State: bgpStateIdle},
			},
			wantCode: gomonitor.Critical,
		},
		{
			name: "all mismatched",
			peers: []BgpPeer{
				{Identifier: "10.0.0.1", AdminStatus: bgpAdminStart, State: bgpStateConnect},
				{Identifier: "10.0.0.2", AdminStatus: bgpAdminStop, State: bgpStateEstablished},
			},
			wantCode: gomonitor.Critical,
		},
		{
			name:     "no peers",
			peers:    nil,
			wantCode: gomonitor.OK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateBgpPeers(tt.peers, false)
			if got.ExitCode != tt.wantCode {
				t.Errorf("got %s (%q), want %s", got.ExitCode, got.Message, tt.wantCode)
			}
			if tt.wantCode == gomonitor.Critical && !strings.Contains(got.Message, "admin status mismatch") {
				t.Errorf("message %q should name the admin status mismatch", got.Message)
			}
		})
	}
}

// TestEvaluateBgpPeers_PerfData verifies that enablePerfData adds total and
// mismatched peer counts to the performance data.
func TestEvaluateBgpPeers_PerfData(t *testing.T) {
	peers := []BgpPeer{
		{Identifier: "10.0.0.1", AdminStatus: bgpAdminStart, State: bgpStateEstablished},
		{Identifier: "10.0.0.2", AdminStatus: bgpAdminStart, State: bgpStateIdle},
	}

	got := EvaluateBgpPeers(peers, true)
	output := got.FormatResult()
	if !strings.Contains(output, "'total_peers'=2") {
		t.Errorf("FormatResult %q does not contain the total_peers metric", output)
	}
	if !strings.Contains(output, "'mismatched_peers'=1") {
		t.Errorf("FormatResult %q does not contain the mismatched_peers metric", output)
	}
}

// TestGetBgpPeers exercises the SNMP walk and per-peer status queries against
// a live device.
func TestGetBgpPeers(t *testing.T) {
	t.Skip("Integration test - requires SNMP device with BGP4-MIB")
}
