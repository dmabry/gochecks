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
	"testing"
)

func TestCollectSystemInfo(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestCollectInterfaces(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestCollectIPAddresses(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestCollectPhysicalEntities(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestCollectCPUMetrics(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestCollectMemoryMetrics(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestParseInterfaceIndex(t *testing.T) {
	tests := []struct {
		name        string
		suffix      string
		want        int
		expectError bool
	}{
		{
			name:        "valid index",
			suffix:      "5",
			want:        5,
			expectError: false,
		},
		{
			name:        "invalid format",
			suffix:      "abc",
			want:        0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInterfaceIndex(tt.suffix)
			if result != tt.want && !tt.expectError {
				t.Errorf("parseInterfaceIndex() = %d, want %d", result, tt.want)
			}
		})
	}
}
