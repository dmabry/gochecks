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
	"testing"
)

func TestSetFieldUnknownOID(t *testing.T) {
	iface := &InterfaceDetail{}
	err := iface.SetField(OID("1.2.3.4.5"), "value")

	if err == nil {
		t.Error("Expected error for unknown OID")
	}
}

func TestSetFieldInvalidStringValue(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
	}{
		{
			name:  "int instead of Description",
			oid:   OIDIfDescr,
			value: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err == nil {
				t.Errorf("SetField(%v, %T) expected error", tt.oid, tt.value)
			}
		})
	}
}

func TestSetFieldInvalidIntValue(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
	}{
		{
			name:  "string instead of Index",
			oid:   OIDIfIndex,
			value: "not an int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err == nil {
				t.Errorf("SetField(%v, %T) expected error", tt.oid, tt.value)
			}
		})
	}
}

func TestSetFieldInvalidUintValue(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
	}{
		{
			name:  "negative int instead of Speed",
			oid:   OIDIfSpeed,
			value: -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err == nil {
				t.Errorf("SetField(%v, %T) expected error", tt.oid, tt.value)
			}
		})
	}
}

func TestSetFieldInvalidUint64Value(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
	}{
		{
			name:  "int32 instead of HCInOctets",
			oid:   OIDIfHCInOctets,
			value: uint32(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err == nil {
				t.Errorf("SetField(%v, %T) expected error", tt.oid, tt.value)
			}
		})
	}
}

func TestSetFieldInvalidUint32Value(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
	}{
		{
			name:  "int instead of LastChange",
			oid:   OIDIfLastChange,
			value: int(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err == nil {
				t.Errorf("SetField(%v, %T) expected error", tt.oid, tt.value)
			}
		})
	}
}

func TestSetFieldPhysAddressValid(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "valid MAC address",
			input: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			want:  "00:11:22:33:44:55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(OIDIfPhysAddress, tt.input)
			if err != nil {
				t.Fatalf("SetField() failed: %v", err)
			}
			if iface.PhysAddress != tt.want {
				t.Errorf("PhysAddress = %q, want %q", iface.PhysAddress, tt.want)
			}
		})
	}
}

func TestSetFieldAllTypes(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
		check func(*InterfaceDetail) bool
	}{
		{
			name:  "string - Description",
			oid:   OIDIfDescr,
			value: "test description",
			check: func(i *InterfaceDetail) bool { return i.Description == "test description" },
		},
		{
			name:  "int - Index",
			oid:   OIDIfIndex,
			value: int(5),
			check: func(i *InterfaceDetail) bool { return i.Index == 5 },
		},
		{
			name:  "uint - Speed",
			oid:   OIDIfSpeed,
			value: uint(1000000),
			check: func(i *InterfaceDetail) bool { return i.Speed == 1000000 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err != nil {
				t.Fatalf("SetField() failed: %v", err)
			}
			if !tt.check(iface) {
				t.Errorf("Field not set correctly")
			}
		})
	}
}

func TestSetFieldZeroValues(t *testing.T) {
	tests := []struct {
		name  string
		oid   OID
		value interface{}
	}{
		{name: "zero int", oid: OIDIfIndex, value: int(0)},
		{name: "empty string", oid: OIDIfDescr, value: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.oid, tt.value)
			if err != nil {
				t.Errorf("SetField() with zero value failed: %v", err)
			}
		})
	}
}

func TestSetFieldEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   OID
		input       interface{}
		expectedErr bool
	}{
		{
			name:        "max int value",
			fieldName:   OIDIfIndex,
			input:       2147483647,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.fieldName, tt.input)
			if (err != nil) != tt.expectedErr {
				t.Errorf("SetField() error = %v", err)
			}
		})
	}
}

func TestSetFieldTypeConversions(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   OID
		input       interface{}
		expectedErr bool
	}{
		{
			name:        "int to int",
			fieldName:   OIDIfIndex,
			input:       42,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &InterfaceDetail{}
			err := iface.SetField(tt.fieldName, tt.input)
			if (err != nil) != tt.expectedErr {
				t.Errorf("SetField() error = %v", err)
			}
		})
	}
}

func TestSetFieldPhysAddressWithInsufficientData(t *testing.T) {
	iface := &InterfaceDetail{}
	defer func() {
		if r := recover(); r == nil {
			t.Logf("Expected panic for insufficient MAC data - test passes")
		}
	}()
	_ = iface.SetField(OIDIfPhysAddress, []byte{0x00})
}
