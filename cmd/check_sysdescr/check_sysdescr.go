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
	"github.com/dmabry/gochecks/internal"
	"github.com/dmabry/gomonitor"
	"regexp"
)

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
	message := fmt.Sprintf("%s", sysDescr)
	checkResult.SetResult(gomonitor.OK, message)

	if enablePerfData {
		// If performance data is enabled, add SNMP latency to the check result
		checkResult.AddPerformanceData("latency", gomonitor.PerformanceMetric{Value: latency.Seconds(), UnitOM: "seconds"})
	}
	return checkResult
}

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
