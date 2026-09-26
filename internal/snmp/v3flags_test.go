package snmp

import (
	"flag"
	"testing"

	"github.com/gosnmp/gosnmp"
)

// TestSNMPVersionFlag verifies the -snmpVersion flag value parsing.
func TestSNMPVersionFlag(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       gosnmp.SnmpVersion
		wantErr    bool
		wantString string
	}{
		{name: "2c", input: "2c", want: gosnmp.Version2c, wantString: "2c"},
		{name: "3", input: "3", want: gosnmp.Version3, wantString: "3"},
		{name: "unsupported v1", input: "1", wantErr: true},
		{name: "garbage", input: "v3", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &SNMPVersionFlag{}
			err := f.Set(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Set(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("Set(%q) unexpected error: %v", tc.input, err)
			}
			if f.Version != tc.want {
				t.Errorf("Set(%q) Version = %v, want %v", tc.input, f.Version, tc.want)
			}
			if got := f.String(); got != tc.wantString {
				t.Errorf("String() = %q, want %q", got, tc.wantString)
			}
		})
	}
}

// TestAuthProtocolFlagAndPrivProtocolFlag verifies protocol flag parsing.
func TestAuthProtocolFlagAndPrivProtocolFlag(t *testing.T) {
	auth := &AuthProtocolFlag{}
	if err := auth.Set("MD5"); err != nil {
		t.Fatalf("auth Set(MD5): %v", err)
	}
	if auth.Protocol != gosnmp.MD5 {
		t.Errorf("auth protocol = %v, want MD5", auth.Protocol)
	}
	if err := auth.Set("bogus"); err == nil {
		t.Error("auth Set(bogus) expected error")
	}

	priv := &PrivProtocolFlag{}
	for _, tc := range []struct {
		input string
		want  gosnmp.SnmpV3PrivProtocol
	}{
		{"DES", gosnmp.DES},
		{"AES", gosnmp.AES},
		{"AES256", gosnmp.AES256},
		{"AES256C", gosnmp.AES256C},
	} {
		if err := priv.Set(tc.input); err != nil {
			t.Fatalf("priv Set(%s): %v", tc.input, err)
		}
		if priv.Protocol != tc.want {
			t.Errorf("priv protocol = %v, want %v", priv.Protocol, tc.want)
		}
	}
	if err := priv.Set("bogus"); err == nil {
		t.Error("priv Set(bogus) expected error")
	}
}

// TestApplyV3Flags_SecurityLevelDerivation verifies that the security level
// is derived from which passphrases were provided.
func TestApplyV3Flags_SecurityLevelDerivation(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		authPassphrase string
		privPassphrase string
		wantLevel      gosnmp.SnmpV3MsgFlags
		wantErr        bool
	}{
		{name: "noAuthNoPriv", username: "user", wantLevel: gosnmp.NoAuthNoPriv},
		{name: "authNoPriv", username: "user", authPassphrase: "authpass", wantLevel: gosnmp.AuthNoPriv},
		{name: "authPriv", username: "user", authPassphrase: "authpass", privPassphrase: "privpass", wantLevel: gosnmp.AuthPriv},
		{name: "priv without auth is invalid", username: "user", privPassphrase: "privpass", wantErr: true},
		{name: "missing username", wantLevel: 0, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{Target: "test"}
			err := ApplyV3Flags(c, tc.username, gosnmp.SHA, tc.authPassphrase, gosnmp.AES, tc.privPassphrase)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (SecLevel=%v)", c.SecLevel)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Version != gosnmp.Version3 {
				t.Errorf("Version = %v, want Version3", c.Version)
			}
			if c.SecLevel != tc.wantLevel {
				t.Errorf("SecLevel = %v, want %v", c.SecLevel, tc.wantLevel)
			}
			if c.SecName != tc.username {
				t.Errorf("SecName = %q, want %q", c.SecName, tc.username)
			}
		})
	}
}

// TestAddV3FlagsRegistersAll verifies AddV3Flags registers every v3 flag on
// the flag set and parses values through a real flag parse.
func TestAddV3FlagsRegistersAll(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	username, authProtocol, authPassphrase, privProtocol, privPassphrase := AddV3Flags(fs)

	args := []string{
		"-v3Username", "monitor",
		"-v3AuthProtocol", "SHA",
		"-v3AuthPassphrase", "authsecret",
		"-v3PrivProtocol", "AES256",
		"-v3PrivPassphrase", "privsecret",
	}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if *username != "monitor" {
		t.Errorf("username = %q, want monitor", *username)
	}
	if authProtocol.Protocol != gosnmp.SHA {
		t.Errorf("auth protocol = %v, want SHA", authProtocol.Protocol)
	}
	if *authPassphrase != "authsecret" {
		t.Errorf("auth passphrase = %q", *authPassphrase)
	}
	if privProtocol.Protocol != gosnmp.AES256 {
		t.Errorf("priv protocol = %v, want AES256", privProtocol.Protocol)
	}
	if *privPassphrase != "privsecret" {
		t.Errorf("priv passphrase = %q", *privPassphrase)
	}
}

// TestNewClientDefaultVersion verifies NewClient defaults to v2c so existing
// v2c behavior is unchanged.
func TestNewClientDefaultVersion(t *testing.T) {
	c, err := NewClient("10.0.0.1", "public")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.Version != gosnmp.Version2c {
		t.Errorf("default Version = %v, want Version2c", c.Version)
	}
}

// TestClientV3OptionsSetFields verifies the WithV3* ClientOptions set the
// USM fields and force Version3.
func TestClientV3OptionsSetFields(t *testing.T) {
	c, err := NewClient("10.0.0.1", "", WithV3AuthPriv("svc", gosnmp.SHA, "ap", gosnmp.AES, "pp"))
	if err != nil {
		t.Fatalf("NewClient with AuthPriv: %v", err)
	}
	if c.Version != gosnmp.Version3 || c.SecLevel != gosnmp.AuthPriv || c.SecName != "svc" {
		t.Errorf("client = %+v", c)
	}
	if c.AuthProtocol != gosnmp.SHA || c.PrivProtocol != gosnmp.AES {
		t.Errorf("protocols = %v/%v, want SHA/AES", c.AuthProtocol, c.PrivProtocol)
	}

	if _, err := NewClient("10.0.0.1", "", WithV3AuthNoPriv("svc", gosnmp.MD5, "ap")); err != nil {
		t.Fatalf("NewClient with AuthNoPriv: %v", err)
	}

	if _, err := NewClient("10.0.0.1", "", WithV3NoAuthNoPriv("svc")); err != nil {
		t.Fatalf("NewClient with NoAuthNoPriv: %v", err)
	}

	// Validation: empty username/passphrase rejected
	if _, err := NewClient("10.0.0.1", "", WithV3NoAuthNoPriv("")); err == nil {
		t.Error("empty username expected error")
	}
	if _, err := NewClient("10.0.0.1", "", WithV3AuthPriv("svc", gosnmp.SHA, "ap", gosnmp.AES, "")); err == nil {
		t.Error("empty priv passphrase expected error")
	}
}

// TestCreateGoSNMPV3Config verifies createGoSNMP configures the gosnmp
// connection for v3 without connecting (validation errors on Connect are
// expected for a nonexistent target; the config is what we check).
func TestCreateGoSNMPV3Config(t *testing.T) {
	c := &Client{
		Target:         "127.0.0.1",
		Version:        gosnmp.Version3,
		SecLevel:       gosnmp.AuthPriv,
		SecName:        "svc",
		AuthProtocol:   gosnmp.SHA,
		AuthPassphrase: "ap",
		PrivProtocol:   gosnmp.AES,
		PrivPassphrase: "pp",
	}
	// createGoSNMP connects to localhost:161; in a sandbox there may be no
	// agent. We only assert that construction does not panic and the
	// gosnmp-level config is applied by inspecting a failed or successful
	// connection result.
	snmpClient, err := c.createGoSNMP()
	if err != nil {
		t.Logf("Connect failed (no local SNMP agent expected in test env): %v", err)
		return
	}
	defer func() {
		if closeErr := snmpClient.Conn.Close(); closeErr != nil {
			t.Logf("error closing SNMP connection: %v", closeErr)
		}
	}()
	if snmpClient.Version != gosnmp.Version3 {
		t.Errorf("Version = %v, want Version3", snmpClient.Version)
	}
	if snmpClient.SecurityModel != gosnmp.UserSecurityModel {
		t.Errorf("SecurityModel = %v, want UserSecurityModel", snmpClient.SecurityModel)
	}
	// gosnmp ORs in the Reportable flag (0x4) for v3 connections, so compare
	// only the security level bits (0x3).
	if snmpClient.MsgFlags&gosnmp.AuthPriv != gosnmp.AuthPriv {
		t.Errorf("MsgFlags = %v, want AuthPriv bits set", snmpClient.MsgFlags)
	}
	usm, ok := snmpClient.SecurityParameters.(*gosnmp.UsmSecurityParameters)
	if !ok {
		t.Fatalf("SecurityParameters is %T, want *UsmSecurityParameters", snmpClient.SecurityParameters)
	}
	if usm.UserName != "svc" || usm.AuthenticationProtocol != gosnmp.SHA || usm.PrivacyProtocol != gosnmp.AES {
		t.Errorf("USM = %+v", usm)
	}
}

// TestVersionOrDefault_UnsetFlagIsV2c verifies the unset -snmpVersion flag
// maps to gosnmp.Version2c, not v1. Regression: an unset flag.Value keeps its
// Go zero value, and Version 0 equals gosnmp.Version1 (0x0), so binaries that
// read Version directly sent SNMPv1 requests (breaking BulkWalk-based checks)
// while the flag's String() helpfully reported "2c".
func TestVersionOrDefault_UnsetFlagIsV2c(t *testing.T) {
	tests := []struct {
		name  string
		input *string // nil means the flag was never set
		want  gosnmp.SnmpVersion
	}{
		{name: "unset flag defaults to 2c", input: nil, want: gosnmp.Version2c},
		{name: "explicit 2c", input: strPtr("2c"), want: gosnmp.Version2c},
		{name: "explicit 3", input: strPtr("3"), want: gosnmp.Version3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &SNMPVersionFlag{}
			if tc.input != nil {
				if err := f.Set(*tc.input); err != nil {
					t.Fatalf("Set(%q): %v", *tc.input, err)
				}
			}
			if got := f.VersionOrDefault(); got != tc.want {
				t.Errorf("VersionOrDefault() = %v, want %v", got, tc.want)
			}
		})
	}

	// Nil receiver must also be safe (flag package can call methods on a nil
	// pointer before Set is ever invoked).
	var f *SNMPVersionFlag
	if got := f.VersionOrDefault(); got != gosnmp.Version2c {
		t.Errorf("nil flag VersionOrDefault() = %v, want Version2c", got)
	}
}

// TestVersionOrDefaultThroughFlagParse verifies the full flag-parse path a
// binary uses: register with flag.Var, parse args WITHOUT -snmpVersion, and
// confirm VersionOrDefault returns v2c.
func TestVersionOrDefaultThroughFlagParse(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	sv := &SNMPVersionFlag{}
	fs.Var(sv, "snmpVersion", "SNMP protocol version.")
	// Mirror the binaries: register the v3 flags on the same flag set.
	snmpUser := fs.String("v3Username", "", "SNMP v3 USM username.")

	// Parse args that omit -snmpVersion entirely, exactly as a v2c deployment
	// invokes the checks.
	if err := fs.Parse([]string{"-v3Username", "monitor"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if *snmpUser != "monitor" {
		t.Fatalf("v3Username = %q, want monitor", *snmpUser)
	}

	if got := sv.VersionOrDefault(); got != gosnmp.Version2c {
		t.Errorf("after parse without -snmpVersion, VersionOrDefault() = %v, want Version2c", got)
	}
}

func strPtr(s string) *string { return &s }
