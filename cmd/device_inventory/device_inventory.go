


package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dmabry/gochecks/internal/snmp"
	"log"
	"strings"
)

type InventoryResult struct {
	SystemInfo      SystemInfo      `json:"system_info,omitempty"`
	Interfaces      []Interface     `json:"interfaces,omitempty"`
	IPAddresses     []IPAddress     `json:"ip_addresses,omitempty"`
	PhysicalEntities []PhysicalEntity `json:"physical_entities,omitempty"`
	CPU             *CPUMetrics     `json:"cpu,omitempty"`
	Memory          *MemoryMetrics  `json:"memory,omitempty"`
}

type SystemInfo struct {
	Description    string `json:"description,omitempty"`
	ObjectID      string `json:"object_id,omitempty"`
	UpTime        float64 `json:"uptime_seconds,omitempty"`
	Contact       string `json:"contact,omitempty"`
	Name          string `json:"name,omitempty"`
	Location      string `json:"location,omitempty"`
}

type Interface struct {
	Index         int    `json:"index,omitempty"`
	Description   string `json:"description,omitempty"`
	Type          int    `json:"type,omitempty"`
	MTU           int    `json:"mtu,omitempty"`
	Speed         int64  `json:"speed_bps,omitempty"`
	MACAddress    string `json:"mac_address,omitempty"`
	AdminStatus   int    `json:"admin_status,omitempty"`
	OperStatus    int    `json:"oper_status,omitempty"`
	InOctets      int64  `json:"in_octets,omitempty"`
	OutOctets     int64  `json:"out_octets,omitempty"`
}

type IPAddress struct {
	IP       string `json:"ip_address,omitempty"`
	IfIndex  int    `json:"interface_index,omitempty"`
}

type PhysicalEntity struct {
	Index        int    `json:"index,omitempty"`
	Description  string `json:"description,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	ModelName    string `json:"model_name,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
}

type CPUMetrics struct {
	User float64 `json:"user_percent,omitempty"`
	System float64 `json:"system_percent,omitempty"`
	Idle float64 `json:"idle_percent,omitempty"`
}

type MemoryMetrics struct {
	TotalSwap int64  `json:"total_swap_kb,omitempty"`
	AvailSwap int64  `json:"avail_swap_kb,omitempty"`
}

// CollectDeviceInventory collects comprehensive inventory information from an SNMP device
func CollectDeviceInventory(snmpClient *snmp.Client) (*InventoryResult, error) {
	result := &InventoryResult{}

	// Collect system information
	systemInfo, err := collectSystemInfo(snmpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to collect system info: %w", err)
	}
	result.SystemInfo = *systemInfo

	// Collect interface information
	interfaces, err := collectInterfaces(snmpClient)
	if err != nil {
		log.Printf("Warning: failed to collect interfaces: %v", err)
	} else {
		result.Interfaces = interfaces
	}

	// Collect IP address information
	ipAddresses, err := collectIPAddresses(snmpClient)
	if err != nil {
		log.Printf("Warning: failed to collect IP addresses: %v", err)
	} else {
		result.IPAddresses = ipAddresses
	}

	// Collect physical entity information
	physicalEntities, err := collectPhysicalEntities(snmpClient)
	if err != nil {
		log.Printf("Warning: failed to collect physical entities: %v", err)
	} else {
		result.PhysicalEntities = physicalEntities
	}

	// Collect CPU metrics (optional)
	cpuMetrics, err := collectCPUMetrics(snmpClient)
	if err != nil {
		log.Printf("Warning: failed to collect CPU metrics: %v", err)
	} else if cpuMetrics != nil {
		result.CPU = cpuMetrics
	}

	// Collect memory metrics (optional)
	memoryMetrics, err := collectMemoryMetrics(snmpClient)
	if err != nil {
		log.Printf("Warning: failed to collect memory metrics: %v", err)
	} else if memoryMetrics != nil {
		result.Memory = memoryMetrics
	}

	return result, nil
}

func collectSystemInfo(client *snmp.Client) (*SystemInfo, error) {
	info := &SystemInfo{}

	oids := []string{
		"1.3.6.1.2.1.1.1.0",  // sysDescr
		"1.3.6.1.2.1.1.2.0",  // sysObjectID
		"1.3.6.1.2.1.1.3.0",  // sysUpTime (in timeticks)
		"1.3.6.1.2.1.1.4.0",  // sysContact
		"1.3.6.1.2.1.1.5.0",  // sysName
		"1.3.6.1.2.1.1.6.0",  // sysLocation
	}

	result, _, err := client.GetValue(oids)
	if err != nil {
		return nil, err
	}

	for i, oid := range oids {
		if i < len(result.Variables) {
			value := result.Variables[i].Value
			switch oid {
			case "1.3.6.1.2.1.1.1.0", "1.3.6.1.2.1.1.2.0", "1.3.6.1.2.1.1.4.0", "1.3.6.1.2.1.1.5.0", "1.3.6.1.2.1.1.6.0":
				if val, ok := value.([]byte); ok {
					switch oid {
					case "1.3.6.1.2.1.1.1.0":
						info.Description = string(val)
					case "1.3.6.1.2.1.1.2.0":
						info.ObjectID = string(val)
					case "1.3.6.1.2.1.1.4.0":
						info.Contact = string(val)
					case "1.3.6.1.2.1.1.5.0":
						info.Name = string(val)
					case "1.3.6.1.2.1.1.6.0":
						info.Location = string(val)
					}
				}
			case "1.3.6.1.2.1.1.3.0": // sysUpTime
				if val, ok := value.(uint32); ok {
					// Convert timeticks to seconds (1 timetick = 1/100 second)
					info.UpTime = float64(val) / 100
				}
			}
		}
	}

	return info, nil
}

func collectInterfaces(client *snmp.Client) ([]Interface, error) {
	var interfaces []Interface

	// Use Walk to get all interface information
	baseOID := "1.3.6.1.2.1.2.2.1"
	oidsMap, _, err := client.Walk(baseOID)
	if err != nil {
		return nil, err
	}

	interfaceDetails := make(map[int]*Interface)

	for oid, value := range oidsMap {
		fields := strings.Split(oid, ".")
		if len(fields) < 2 {
			continue
		}

		indexStr := fields[len(fields)-1]
		index := parseInterfaceIndex(indexStr)
		if index == 0 {
			continue
		}

		if _, ok := interfaceDetails[index]; !ok {
			interfaceDetails[index] = &Interface{Index: index}
		}

		iface := interfaceDetails[index]

		switch oid {
		case "1.3.6.1.2.1.2.2.1.1": // ifIndex
			if val, ok := value.(int); ok {
				iface.Index = val
			}
		case "1.3.6.1.2.1.2.2.1.2": // ifDescr
			if val, ok := value.([]byte); ok {
				iface.Description = string(val)
			}
		case "1.3.6.1.2.1.2.2.1.3": // ifType
			if val, ok := value.(int); ok {
				iface.Type = val
			}
		case "1.3.6.1.2.1.2.2.1.4": // ifMtu
			if val, ok := value.(int); ok {
				iface.MTU = val
			}
		case "1.3.6.1.2.1.2.2.1.5": // ifSpeed
			if val, ok := value.(uint64); ok {
				iface.Speed = int64(val)
			} else if val, ok := value.(int); ok {
				iface.Speed = int64(val)
			}
		case "1.3.6.1.2.1.2.2.1.6": // ifPhysAddress
			if val, ok := value.([]byte); ok && len(val) == 6 {
				iface.MACAddress = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
					val[0], val[1], val[2], val[3], val[4], val[5])
			}
		case "1.3.6.1.2.1.2.2.1.7": // ifAdminStatus
			if val, ok := value.(int); ok {
				iface.AdminStatus = val
			}
		case "1.3.6.1.2.1.2.2.1.8": // ifOperStatus
			if val, ok := value.(int); ok {
				iface.OperStatus = val
			}
		case "1.3.6.1.2.1.2.2.1.10": // ifInOctets
			if val, ok := value.(uint64); ok {
				iface.InOctets = int64(val)
			} else if val, ok := value.(int); ok {
				iface.InOctets = int64(val)
			}
		case "1.3.6.1.2.1.2.2.1.16": // ifOutOctets
			if val, ok := value.(uint64); ok {
				iface.OutOctets = int64(val)
			} else if val, ok := value.(int); ok {
				iface.OutOctets = int64(val)
			}
		}
	}

	for _, iface := range interfaceDetails {
		interfaces = append(interfaces, *iface)
	}

	return interfaces, nil
}

// parseInterfaceIndex extracts the interface index from an OID suffix
func parseInterfaceIndex(suffix string) int {
	index := 0
	fmt.Sscanf(suffix, "%d", &index)
	return index
}

var currentIfIndex int

func collectIPAddresses(client *snmp.Client) ([]IPAddress, error) {
	var ipAddresses []IPAddress

	// Use Walk to get IP address table
	baseOID := "1.3.6.1.2.1.4.20.1"
	oidsMap, _, err := client.Walk(baseOID)
	if err != nil {
		return nil, err
	}

	for oid, value := range oidsMap {
		switch oid {
		case "1.3.6.1.2.1.4.20.1.1": // ipAdEntIfIndex
			if val, ok := value.(int); ok {
				currentIfIndex = val
			}
		case "1.3.6.1.2.1.4.20.1.2": // ipAdEntAddr (IP address)
			if val, ok := value.([]byte); ok && len(val) == 4 {
				ipInfo := IPAddress{
					IfIndex: currentIfIndex,
					IP:      fmt.Sprintf("%d.%d.%d.%d", val[0], val[1], val[2], val[3]),
				}
				ipAddresses = append(ipAddresses, ipInfo)
			}
		}
	}

	return ipAddresses, nil
}

var currentEntityIndex int

func collectPhysicalEntities(client *snmp.Client) ([]PhysicalEntity, error) {
	var entities []PhysicalEntity

	// Use Walk to get physical entity table
	baseOID := "1.3.6.1.2.1.47.1.1.1.1"
	oidsMap, _, err := client.Walk(baseOID)
	if err != nil {
		return nil, err
	}

	currentEntityIndex = 0

	for oid, value := range oidsMap {
		switch oid {
		case "1.3.6.1.2.1.47.1.1.1.1.2": // entPhysicalDescr
			if val, ok := value.([]byte); ok {
				currentEntityIndex++
				currentEntity := PhysicalEntity{
					Index:       currentEntityIndex,
					Description: string(val),
				}
				entities = append(entities, currentEntity)
			}
		case "1.3.6.1.2.1.47.1.1.1.1.3": // entPhysicalVendor
			if len(entities) > 0 {
				if val, ok := value.([]byte); ok {
					entities[len(entities)-1].Vendor = string(val)
				}
			}
		case "1.3.6.1.2.1.47.1.1.1.1.5": // entPhysicalModelName
			if len(entities) > 0 {
				if val, ok := value.([]byte); ok {
					entities[len(entities)-1].ModelName = string(val)
				}
			}
		case "1.3.6.1.2.1.47.1.1.1.1.11": // entPhysicalSerialNum
			if len(entities) > 0 {
				if val, ok := value.([]byte); ok {
					entities[len(entities)-1].SerialNumber = string(val)
				}
			}
		}
	}

	return entities, nil
}

func collectCPUMetrics(client *snmp.Client) (*CPUMetrics, error) {
	metrics := &CPUMetrics{}

	// Use UCD-SNMP-MIB for CPU metrics (Linux/Unix systems)
	oids := []string{
		"1.3.6.1.4.1.2021.11.50.0", // ssCpuRawUser
		"1.3.6.1.4.1.2021.11.51.0", // ssCpuRawSystem
		"1.3.6.1.4.1.2021.11.52.0", // ssCpuRawIdle
	}

	result, _, err := client.GetValue(oids)
	if err != nil {
		return nil, err
	}

	totalTicks := float64(0)
	for i, oid := range oids {
		if i < len(result.Variables) {
			value := result.Variables[i].Value
			switch oid {
			case "1.3.6.1.4.1.2021.11.50.0", "1.3.6.1.4.1.2021.11.51.0", "1.3.6.1.4.1.2021.11.52.0":
				if val, ok := value.(uint32); ok {
					switch oid {
					case "1.3.6.1.4.1.2021.11.50.0":
						metrics.User = float64(val)
					case "1.3.6.1.4.1.2021.11.51.0":
						metrics.System = float64(val)
					case "1.3.6.1.4.1.2021.11.52.0":
						metrics.Idle = float64(val)
					}
					totalTicks += float64(val)
				}
			}
		}
	}

	if totalTicks > 0 {
		metrics.User = (metrics.User / totalTicks) * 100
		metrics.System = (metrics.System / totalTicks) * 100
		metrics.Idle = (metrics.Idle / totalTicks) * 100
		return metrics, nil
	}

	return nil, nil // Not available on this device
}

func collectMemoryMetrics(client *snmp.Client) (*MemoryMetrics, error) {
	metrics := &MemoryMetrics{}

	// Use UCD-SNMP-MIB for memory metrics (Linux/Unix systems)
	oids := []string{
		"1.3.6.1.4.1.2021.4.3.0", // memTotalSwap
		"1.3.6.1.4.1.2021.4.4.0", // memAvailSwap
	}

	result, _, err := client.GetValue(oids)
	if err != nil {
		return nil, err
	}

	for i, oid := range oids {
		if i < len(result.Variables) {
			value := result.Variables[i].Value
			switch oid {
			case "1.3.6.1.4.1.2021.4.3.0", "1.3.6.1.4.1.2021.4.4.0":
				if val, ok := value.(int); ok {
					switch oid {
					case "1.3.6.1.4.1.2021.4.3.0":
						metrics.TotalSwap = int64(val)
					case "1.3.6.1.4.1.2021.4.4.0":
						metrics.AvailSwap = int64(val)
					}
				}
			}
		}
	}

	if metrics.TotalSwap > 0 && metrics.AvailSwap >= 0 {
		return metrics, nil
	}

	return nil, nil // Not available on this device
}

// main is the entry point of the program. It parses command-line flags and collects device inventory.
func main() {
	target := flag.String("target", "127.0.0.1", "The target SNMP device.")
	community := flag.String("community", "public", "The SNMP community string.")
	outputFormat := flag.String("output", "json", "Output format (currently only \"json\" is supported)")
	flag.Parse()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
	}

	result, err := CollectDeviceInventory(&snmpClient)
	if err != nil {
		log.Fatalf("Error collecting inventory: %v", err)
	}

	// Output the result in the requested format
	switch *outputFormat {
	case "json":
		outputJSON(result)
	default:
		log.Fatalf("Unsupported output format: %s", *outputFormat)
	}
}

// outputJSON prints the inventory result as JSON
func outputJSON(result *InventoryResult) {
	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	fmt.Println(string(jsonOutput))
}


