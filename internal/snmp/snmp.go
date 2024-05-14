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
	"github.com/gosnmp/gosnmp"
	"time"
)

// timeout15 is a constant representing a timeout duration of 15 seconds.
const (
	timeout15 = time.Duration(15) * time.Second
)

// Client represents an SNMP client that allows connecting to a target SNMP device.
type Client struct {
	Target    string
	Community string
}

// Connect establishes a connection to the SNMP target using the provided parameters,
// and returns a GoSNMP client instance along with any error encountered during connection.
// The function sets the default SNMP port to 161 and the SNMP version to 2c.
// The function also sets the timeout duration to 15 seconds.
// If an error occurs while connecting to the target, nil is returned along with the error.
//
// Example usage:
// snmpClient, err := client.Connect()
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// defer snmpClient.Conn.Close()
// ...
func (s *Client) Connect() (*gosnmp.GoSNMP, error) {
	snmpClient := &gosnmp.GoSNMP{
		Target:    s.Target,
		Port:      161,
		Community: s.Community,
		Version:   gosnmp.Version2c,
		Timeout:   timeout15,
	}

	if err := snmpClient.Connect(); err != nil {
		return nil, err
	}

	return snmpClient, nil
}

// GetValue retrieves SNMP values for the given OIDs using the client's connection.
// It returns the SNMP packet containing the result values, the duration of the SNMP request,
// and any error encountered during the process.
func (s *Client) GetValue(oids []string) (*gosnmp.SnmpPacket, time.Duration, error) {
	snmpClient, err := s.Connect()
	if err != nil {
		return nil, 0, err
	}
	defer snmpClient.Conn.Close()

	start := time.Now()

	result, err := snmpClient.Get(oids)
	if err != nil {
		return nil, 0, err
	}

	latency := time.Since(start)

	return result, latency, nil
}

// Walk retrieves SNMP tree for the given OID using the client's connection.
// It returns a map with the OID as the key and its value as the value,
// the duration of the SNMP request, and any error encountered during the process.
func (s *Client) Walk(baseOid string) (map[string]interface{}, time.Duration, error) {
	snmpClient, err := s.Connect()
	if err != nil {
		return nil, 0, err
	}
	defer snmpClient.Conn.Close()

	start := time.Now()

	oidValues := make(map[string]interface{})

	err = snmpClient.BulkWalk(baseOid, func(pdu gosnmp.SnmpPDU) error {
		oidValues[pdu.Name] = pdu.Value
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	latency := time.Since(start)

	return oidValues, latency, nil
}
