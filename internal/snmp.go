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

const (
	timeout15 = time.Duration(15) * time.Second
)

type Client struct {
	Target    string
	Community string
}

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
