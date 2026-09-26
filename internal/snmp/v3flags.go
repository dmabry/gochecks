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
	"flag"
	"fmt"

	"github.com/gosnmp/gosnmp"
)

// SNMPVersionFlag is a flag.Value implementation for SNMP protocol versions,
// accepting "2c" and "3".
type SNMPVersionFlag struct {
	Version gosnmp.SnmpVersion
}

// String implements flag.Value.
func (v *SNMPVersionFlag) String() string {
	if v == nil || v.Version == 0 {
		return "2c"
	}
	switch v.Version {
	case gosnmp.Version1:
		return "1"
	case gosnmp.Version2c:
		return "2c"
	case gosnmp.Version3:
		return "3"
	default:
		return fmt.Sprintf("SnmpVersion(%d)", int(v.Version))
	}
}

// Set implements flag.Value. Only "2c" and "3" are supported.
func (v *SNMPVersionFlag) Set(s string) error {
	switch s {
	case "2c":
		v.Version = gosnmp.Version2c
	case "3":
		v.Version = gosnmp.Version3
	default:
		return fmt.Errorf("unsupported SNMP version %q: only \"2c\" and \"3\" are supported", s)
	}
	return nil
}

// VersionOrDefault returns the parsed SNMP version, mapping the zero value
// left by an unset flag to gosnmp.Version2c. Without this mapping, an unset
// -snmpVersion leaves Version == 0, which equals gosnmp.Version1 (0x0), so
// clients would silently send SNMPv1 requests instead of the documented v2c
// default. Call this instead of reading Version directly when constructing
// SNMP clients from CLI flags.
func (v *SNMPVersionFlag) VersionOrDefault() gosnmp.SnmpVersion {
	if v == nil || v.Version == 0 {
		return gosnmp.Version2c
	}
	return v.Version
}

// AddV3Flags registers SNMP v3 USM flags (username, auth protocol and
// passphrase, priv protocol and passphrase) on the flag set. Use with
// -snmpVersion 3.
func AddV3Flags(fs *flag.FlagSet) (username *string, authProtocol *AuthProtocolFlag, authPassphrase *string, privProtocol *PrivProtocolFlag, privPassphrase *string) {
	username = fs.String("v3Username", "", "SNMP v3 USM username (required for v3).")
	authPassphrase = fs.String("v3AuthPassphrase", "", "SNMP v3 authentication passphrase (required for authNoPriv and authPriv).")
	privPassphrase = fs.String("v3PrivPassphrase", "", "SNMP v3 privacy passphrase (required for authPriv).")
	authProtocol = &AuthProtocolFlag{Protocol: gosnmp.SHA}
	fs.Var(authProtocol, "v3AuthProtocol", "SNMP v3 authentication protocol: MD5 or SHA (default SHA).")
	privProtocol = &PrivProtocolFlag{Protocol: gosnmp.AES}
	fs.Var(privProtocol, "v3PrivProtocol", "SNMP v3 privacy protocol: DES, AES, AES192, AES256, AES192C, or AES256C (default AES).")
	return username, authProtocol, authPassphrase, privProtocol, privPassphrase
}

// AuthProtocolFlag is a flag.Value implementation for SNMP v3 auth protocols.
type AuthProtocolFlag struct {
	Protocol gosnmp.SnmpV3AuthProtocol
}

// String implements flag.Value.
func (a *AuthProtocolFlag) String() string {
	if a == nil || a.Protocol == 0 {
		return "SHA"
	}
	switch a.Protocol {
	case gosnmp.MD5:
		return "MD5"
	case gosnmp.SHA:
		return "SHA"
	default:
		return fmt.Sprintf("SnmpV3AuthProtocol(%d)", int(a.Protocol))
	}
}

// Set implements flag.Value.
func (a *AuthProtocolFlag) Set(s string) error {
	switch s {
	case "MD5":
		a.Protocol = gosnmp.MD5
	case "SHA":
		a.Protocol = gosnmp.SHA
	default:
		return fmt.Errorf("unsupported auth protocol %q: only MD5 and SHA are supported", s)
	}
	return nil
}

// PrivProtocolFlag is a flag.Value implementation for SNMP v3 priv protocols.
type PrivProtocolFlag struct {
	Protocol gosnmp.SnmpV3PrivProtocol
}

// String implements flag.Value.
func (p *PrivProtocolFlag) String() string {
	if p == nil || p.Protocol == 0 {
		return "AES"
	}
	switch p.Protocol {
	case gosnmp.DES:
		return "DES"
	case gosnmp.AES:
		return "AES"
	case gosnmp.AES192:
		return "AES192"
	case gosnmp.AES256:
		return "AES256"
	case gosnmp.AES192C:
		return "AES192C"
	case gosnmp.AES256C:
		return "AES256C"
	default:
		return fmt.Sprintf("SnmpV3PrivProtocol(%d)", int(p.Protocol))
	}
}

// Set implements flag.Value.
func (p *PrivProtocolFlag) Set(s string) error {
	switch s {
	case "DES":
		p.Protocol = gosnmp.DES
	case "AES":
		p.Protocol = gosnmp.AES
	case "AES192":
		p.Protocol = gosnmp.AES192
	case "AES256":
		p.Protocol = gosnmp.AES256
	case "AES192C":
		p.Protocol = gosnmp.AES192C
	case "AES256C":
		p.Protocol = gosnmp.AES256C
	default:
		return fmt.Errorf("unsupported priv protocol %q", s)
	}
	return nil
}

// ApplyV3Flags configures a Client for SNMP v3 from parsed CLI flag values.
// The security level is derived from which passphrases were provided:
//   - no passphrases: noAuthNoPriv (username only)
//   - auth passphrase only: authNoPriv
//   - both passphrases: authPriv
func ApplyV3Flags(c *Client, username string, authProtocol gosnmp.SnmpV3AuthProtocol, authPassphrase string, privProtocol gosnmp.SnmpV3PrivProtocol, privPassphrase string) error {
	switch {
	case authPassphrase == "" && privPassphrase == "":
		return c.WithV3NoAuthNoPriv(username)
	case authPassphrase != "" && privPassphrase == "":
		return c.WithV3AuthNoPriv(username, authProtocol, authPassphrase)
	case authPassphrase != "" && privPassphrase != "":
		return c.WithV3AuthPriv(username, authProtocol, authPassphrase, privProtocol, privPassphrase)
	default:
		return fmt.Errorf("invalid v3 configuration: authPriv requires an auth passphrase (-v3AuthPassphrase)")
	}
}

// WithV3NoAuthNoPriv is a pointer-receiver wrapper so ApplyV3Flags can be
// called on a Client value obtained as a pointer.
func (c *Client) WithV3NoAuthNoPriv(username string) error {
	opt := WithV3NoAuthNoPriv(username)
	return opt(c)
}

// WithV3AuthNoPriv is a pointer-receiver wrapper for use by ApplyV3Flags.
func (c *Client) WithV3AuthNoPriv(username string, authProtocol gosnmp.SnmpV3AuthProtocol, authPassphrase string) error {
	opt := WithV3AuthNoPriv(username, authProtocol, authPassphrase)
	return opt(c)
}

// WithV3AuthPriv is a pointer-receiver wrapper for use by ApplyV3Flags.
func (c *Client) WithV3AuthPriv(username string, authProtocol gosnmp.SnmpV3AuthProtocol, authPassphrase string, privProtocol gosnmp.SnmpV3PrivProtocol, privPassphrase string) error {
	opt := WithV3AuthPriv(username, authProtocol, authPassphrase, privProtocol, privPassphrase)
	return opt(c)
}
