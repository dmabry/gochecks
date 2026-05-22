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

	"github.com/dmabry/gochecks/internal/interfaces"
)

func TestUpdateInterfaceDetails(t *testing.T) {
	tests := []struct {
		name          string
		ifaceDetail   *interfaces.InterfaceDetail
		oid           interfaces.OID
		value         interface{}
		expectNoPanic bool
	}{
		{
			name: "valid int value",
			ifaceDetail: &interfaces.InterfaceDetail{
				Index: 0,
			},
			oid:           interfaces.OIDIfIndex,
			value:         1,
			expectNoPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectNoPanic {
				defer func() {
					if r := recover(); r != nil {
						return
					}
				}()
			}
			updateInterfaceDetails(tt.ifaceDetail, tt.oid, tt.value)
		})
	}
}

func TestBuildInterfaceDetailsMessage(t *testing.T) {
	tests := []struct {
		name        string
		interfaces  map[int]*interfaces.InterfaceDetail
		expectEmpty bool
	}{
		{
			name:        "empty map",
			interfaces:  make(map[int]*interfaces.InterfaceDetail),
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildInterfaceDetailsMessage(tt.interfaces)
			if (result == "") != tt.expectEmpty {
				t.Errorf("buildInterfaceDetailsMessage() = %q, expect empty: %v", result, tt.expectEmpty)
			}
		})
	}
}

func TestCheckInterfaceMetrics(t *testing.T) {
	t.Skip("Integration test - requires SNMP device")
}

func TestFilterInterfacesByDescription(t *testing.T) {
	t.Skip("Requires gomonitor mock implementation")
}
