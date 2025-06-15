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
	"flag"
	"fmt"
	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/dmabry/gomonitor"
	"regexp"
)

// CheckSysDescr checks the sysDescr value of an SNMP target using a regular expression pattern.
// It takes the SNMP client, the expected sysDescr regular expression pattern, and a boolean flag to enable performance data.
// It returns a CheckResult struct with the result of the check and the performance data (if enabled).
//
// The function retrieves the sysDescr value using the GetValue method of the SNMP client.
// If an error occurs while retrieving the value, a critical check result is returned with an error message.
//
// If the expectedSysDescrRegExp is provided, the function compares the sysDescr value with the regular expression pattern.
// If it does not match, a critical check result is returned with an error message.
//
// Otherwise, an OK check result is returned with the sysDescr value.
//
// If enablePerfData is true, the function adds the SNMP latency to the performance data of the check result.
// The latency is measured as the duration of the SNMP request.
//
// Example usage:
//
//	snmpClient := snmp.Client{
//	    Target:    "127.0.0.1",
//	    Community: "public",
//	}
//	result := CheckSysDescr(&snmpClient, "Cisco", true)
//	result.SendResult()
func CheckSysDescr(snmpClient *snmp.Client, expectedSysDescrRegExp string, enablePerfData bool) *gomonitor.CheckResult {
	oids := []string{"1.3.6.1.2.1.1.1.0"}

	result, latency, err := snmpClient.GetValue(oids)
	if err != nil {
		checkResult := gomonitor.NewCheckResult()
		eMessage := fmt.Sprintf("SNMP target %s failed to return data for requested OID.", snmpClient.Target)
		checkResult.SetResult(gomonitor.Critical, eMessage)
		return checkResult
	}

	checkResult := gomonitor.NewCheckResult()
	sysDescr := string(result.Variables[0].Value.([]uint8))

	// Compare result with expected sysDescr using regexp
	if expectedSysDescrRegExp != "" {
		match, err := regexp.MatchString(expectedSysDescrRegExp, sysDescr)
		if err != nil || !match {
			eMessage := fmt.Sprintf("sysDescr does not match expected pattern '%s'. Got: %s", expectedSysDescrRegExp, sysDescr)
			checkResult.SetResult(gomonitor.Critical, eMessage)
			return checkResult
		}
	}
	message := sysDescr
	checkResult.SetResult(gomonitor.OK, message)

	if enablePerfData {
		// If performance data is enabled, add SNMP latency to the check result
		checkResult.AddPerformanceData("latency", gomonitor.PerformanceMetric{Value: latency.Seconds(), UnitOM: "seconds"})
	}
	return checkResult
}

// main is the entry point of the program. It parses command-line flags, creates an SNMP client,
// and performs a check on the target SNMP device using the CheckSysDescr function.
// The result of the check is then sent using the SendResult method.
func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	expectedSysDescrRegExp := flag.String("sysDescrPattern", "", "Regex pattern sysDescr to be matched. If not provided, any sysDescr will be accepted.")
	enablePerfData := flag.Bool("enablePerfData", false, "Enable performance data. Default is false.")
	flag.Parse()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
	}
	result := CheckSysDescr(&snmpClient, *expectedSysDescrRegExp, *enablePerfData)
	result.SendResult()
}
