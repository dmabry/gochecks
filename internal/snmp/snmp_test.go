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

package snmp

import (
	"testing"
)

func TestClientConnect(t *testing.T) {
	t.Skip("Requires SNMP device - not a unit test")
	tests := []struct {
		name      string
		target    string
		community string
		wantError bool
	}{
		{
			name:      "valid connection",
			target:    "127.0.0.1",
			community: "public",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Target:    tt.target,
				Community: tt.community,
			}
			err := client.Connect(nil)
			if (err != nil) != tt.wantError {
				t.Errorf("Connect() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestClientGetValue(t *testing.T) {
	t.Skip("Requires SNMP device - not a unit test")
	tests := []struct {
		name      string
		target    string
		community string
		oids      []string
		wantError bool
	}{
		{
			name:      "invalid OID returns error",
			target:    "127.0.0.1",
			community: "public",
			oids:      []string{"1.3.6.9999"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Target:    tt.target,
				Community: tt.community,
			}
			_, _, err := client.GetValue(nil, tt.oids)
			if (err != nil) != tt.wantError {
				t.Errorf("GetValue() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestClientWalk(t *testing.T) {
	t.Skip("Requires SNMP device - not a unit test")
	tests := []struct {
		name       string
		target     string
		community  string
		baseOID    string
		wantError  bool
		minResults int
	}{
		{
			name:       "valid walk returns results",
			target:     "127.0.0.1",
			community:  "public",
			baseOID:    ".1.3.6.1.2.1.1",
			wantError:  false,
			minResults: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Target:    tt.target,
				Community: tt.community,
			}
			result, _, err := client.Walk(nil, tt.baseOID)
			if (err != nil) != tt.wantError {
				t.Errorf("Walk() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) < tt.minResults {
				t.Errorf("Walk() returned %d results, want at least %d", len(result), tt.minResults)
			}
		})
	}
}

func TestTimeoutConstants(t *testing.T) {
	expectedDuration := 15.0
	if timeout15.Seconds() != expectedDuration {
		t.Errorf("timeout15 = %v, want %v seconds", timeout15, expectedDuration)
	}
}

func TestMockClientWalkResultsSetCorrectly(t *testing.T) {
	t.Skip("Requires SNMP device - not a unit test")
}
