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
	"net"
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

// TestApplyInterfaceOID verifies the ifTable column mapping with the native
// value types gosnmp returns: []byte for octet strings, int for Integer
// objects, and uint for Counter32/Gauge32 objects.
func TestApplyInterfaceOID(t *testing.T) {
	tests := []struct {
		name  string
		oid   string
		value interface{}
		check func(*Interface) bool
	}{
		{
			name:  "ifDescr octet string",
			oid:   ".1.3.6.1.2.1.2.2.1.2",
			value: []byte("eth0"),
			check: func(i *Interface) bool { return i.Description == "eth0" },
		},
		{
			name:  "ifSpeed gauge32 as uint",
			oid:   ".1.3.6.1.2.1.2.2.1.5",
			value: uint(1000000000),
			check: func(i *Interface) bool { return i.Speed == 1000000000 },
		},
		{
			name:  "ifInOctets counter32 as uint",
			oid:   ".1.3.6.1.2.1.2.2.1.10",
			value: uint(12345),
			check: func(i *Interface) bool { return i.InOctets == 12345 },
		},
		{
			name:  "ifOutOctets counter32 as uint",
			oid:   ".1.3.6.1.2.1.2.2.1.16",
			value: uint(54321),
			check: func(i *Interface) bool { return i.OutOctets == 54321 },
		},
		{
			name:  "ifPhysAddress mac",
			oid:   ".1.3.6.1.2.1.2.2.1.6",
			value: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			check: func(i *Interface) bool { return i.MACAddress == "00:11:22:33:44:55" },
		},
		{
			name:  "ifAdminStatus integer",
			oid:   ".1.3.6.1.2.1.2.2.1.7",
			value: 1,
			check: func(i *Interface) bool { return i.AdminStatus == 1 },
		},
		{
			name:  "ifOperStatus integer",
			oid:   ".1.3.6.1.2.1.2.2.1.8",
			value: 1,
			check: func(i *Interface) bool { return i.OperStatus == 1 },
		},
		{
			name:  "unknown oid ignored",
			oid:   ".1.3.6.1.2.1.2.2.1.99",
			value: 1,
			check: func(i *Interface) bool { return i.InOctets == 0 && i.Description == "" },
		},
		{
			name:  "wrong type ignored",
			oid:   ".1.3.6.1.2.1.2.2.1.8",
			value: "up",
			check: func(i *Interface) bool { return i.OperStatus == 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iface := &Interface{Index: 1}
			applyInterfaceOID(iface, tt.oid, tt.value)
			if !tt.check(iface) {
				t.Errorf("applyInterfaceOID(%s, %T) field not set as expected", tt.oid, tt.value)
			}
		})
	}
}

// TestApplyEntityOID pins the RFC 6933 entPhysicalTable column numbers:
// descr=.1, serialNum=.10, mfgName=.11, modelName=.12.
func TestApplyEntityOID(t *testing.T) {
	entity := &PhysicalEntity{Index: 1}
	applyEntityOID(entity, ".1.3.6.1.2.1.47.1.1.1.1.1", []byte("Chassis 1"))
	applyEntityOID(entity, ".1.3.6.1.2.1.47.1.1.1.1.10", []byte("SN12345"))
	applyEntityOID(entity, ".1.3.6.1.2.1.47.1.1.1.1.11", []byte("Acme"))
	applyEntityOID(entity, ".1.3.6.1.2.1.47.1.1.1.1.12", []byte("Model-X"))
	applyEntityOID(entity, ".1.3.6.1.2.1.47.1.1.1.1.2", []byte("should not map"))

	if entity.Description != "Chassis 1" {
		t.Errorf("Description = %q, want %q", entity.Description, "Chassis 1")
	}
	if entity.SerialNumber != "SN12345" {
		t.Errorf("SerialNumber = %q, want %q", entity.SerialNumber, "SN12345")
	}
	if entity.Vendor != "Acme" {
		t.Errorf("Vendor = %q, want %q", entity.Vendor, "Acme")
	}
	if entity.ModelName != "Model-X" {
		t.Errorf("ModelName = %q, want %q", entity.ModelName, "Model-X")
	}
}

func TestSplitIPTableOID(t *testing.T) {
	tests := []struct {
		name string
		oid  string
		base string
		ip   string
		ok   bool
	}{
		{
			name: "ipAdEntIfIndex entry",
			oid:  ".1.3.6.1.2.1.4.20.1.2.10.0.0.1",
			base: ".1.3.6.1.2.1.4.20.1.2",
			ip:   "10.0.0.1",
			ok:   true,
		},
		{
			name: "ipAdEntAddr entry",
			oid:  ".1.3.6.1.2.1.4.20.1.1.192.168.1.5",
			base: ".1.3.6.1.2.1.4.20.1.1",
			ip:   "192.168.1.5",
			ok:   true,
		},
		{
			name: "too short",
			oid:  ".1",
			base: "",
			ip:   "",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, ip, ok := splitIPTableOID(tt.oid)
			if ok != tt.ok || base != tt.base || ip != tt.ip {
				t.Errorf("splitIPTableOID(%s) = (%q, %q, %v), want (%q, %q, %v)", tt.oid, base, ip, ok, tt.base, tt.ip, tt.ok)
			}
		})
	}
}

func TestAsIPAddress(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
		ok    bool
	}{
		{name: "gosnmp string", value: "10.0.0.1", want: "10.0.0.1", ok: true},
		{name: "byte slice", value: []byte{10, 0, 0, 1}, want: "10.0.0.1", ok: true},
		{name: "net.IP", value: net.IP{10, 0, 0, 1}, want: "10.0.0.1", ok: true},
		{name: "int rejected", value: 10, want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := asIPAddress(tt.value)
			if ok != tt.ok || got != tt.want {
				t.Errorf("asIPAddress(%T) = (%q, %v), want (%q, %v)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestAsCounter(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  int64
	}{
		{name: "uint (gosnmp Counter32/Gauge32)", value: uint(5), want: 5},
		{name: "uint64", value: uint64(5), want: 5},
		{name: "int", value: 5, want: 5},
		{name: "negative int rejected", value: -5, want: 0},
		{name: "string rejected", value: "5", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asCounter(tt.value); got != tt.want {
				t.Errorf("asCounter(%T) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestAsCounter32(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  float64
		ok    bool
	}{
		{name: "uint (gosnmp Counter32)", value: uint(42), want: 42, ok: true},
		{name: "uint32", value: uint32(42), want: 42, ok: true},
		{name: "string rejected", value: "42", want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := asCounter32(tt.value)
			if ok != tt.ok || got != tt.want {
				t.Errorf("asCounter32(%T) = (%v, %v), want (%v, %v)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestAsOctetString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
		ok    bool
	}{
		{name: "gosnmp []byte", value: []byte("eth0"), want: "eth0", ok: true},
		{name: "plain string", value: "eth0", want: "eth0", ok: true},
		{name: "int rejected", value: 1, want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := asOctetString(tt.value)
			if ok != tt.ok || got != tt.want {
				t.Errorf("asOctetString(%T) = (%q, %v), want (%q, %v)", tt.value, got, ok, tt.want, tt.ok)
			}
		})
	}
}
