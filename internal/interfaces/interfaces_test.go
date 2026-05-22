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
	"testing"
)

func TestInterfaceDetail(t *testing.T) {
	tests := []struct {
		name           string
		ifaceDetail    InterfaceDetail
		expectedOutput string
	}{
		{
			name: "basic interface",
			ifaceDetail: InterfaceDetail{
				Description: "eth0",
				Name:        "eth0",
				Alias:       "",
				PhysAddress: "00:11:22:33:44:55",
				Index:       1,
				Type:        6,
				MTU:         1500,
				Speed:       1000000000,
				HighSpeed:   1000,
				OperStatus:  1,
				AdminStatus: 1,
				InOctets:    1000,
				OutOctets:   500,
			},
			expectedOutput: "Interface index: 1\nDescription: eth0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ifaceDetail.ToString(1)
			if len(result) == 0 {
				t.Error("ToString() returned empty string")
			}
			if result != "" && !contains(result, "Interface index:") {
				t.Error("ToString() should contain 'Interface index:'")
			}
		})
	}
}

func TestInterfaceDetailJSON(t *testing.T) {
	tests := []struct {
		name        string
		ifaceDetail InterfaceDetail
		expectValid bool
	}{
		{
			name:        "valid JSON output",
			ifaceDetail: InterfaceDetail{Index: 1, Name: "test0"},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.ifaceDetail.ToJsonString()
			if (err == nil) != tt.expectValid {
				t.Errorf("ToJsonString() error = %v, expectValid %v", err, tt.expectValid)
			}
			if !tt.expectValid && result != "" {
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(result), &m); err != nil {
					t.Errorf("ToJsonString() returned invalid JSON: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
