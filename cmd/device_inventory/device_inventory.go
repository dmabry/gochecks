package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/dmabry/gochecks/internal/interfaces"
	"github.com/dmabry/gochecks/internal/snmp"
	"github.com/gosnmp/gosnmp"
)

type InventoryResult struct {
	SystemInfo       SystemInfo       `json:"system_info,omitempty"`
	Interfaces       []Interface      `json:"interfaces,omitempty"`
	IPAddresses      []IPAddress      `json:"ip_addresses,omitempty"`
	PhysicalEntities []PhysicalEntity `json:"physical_entities,omitempty"`
	CPU              *CPUMetrics      `json:"cpu,omitempty"`
	Memory           *MemoryMetrics   `json:"memory,omitempty"`
}

type SystemInfo struct {
	Description string  `json:"description,omitempty"`
	ObjectID    string  `json:"object_id,omitempty"`
	UpTime      float64 `json:"uptime_seconds,omitempty"`
	Contact     string  `json:"contact,omitempty"`
	Name        string  `json:"name,omitempty"`
	Location    string  `json:"location,omitempty"`
}

type Interface struct {
	Index       int    `json:"index,omitempty"`
	Description string `json:"description,omitempty"`
	Type        int    `json:"type,omitempty"`
	MTU         int    `json:"mtu,omitempty"`
	Speed       int64  `json:"speed_bps,omitempty"`
	MACAddress  string `json:"mac_address,omitempty"`
	AdminStatus int    `json:"admin_status,omitempty"`
	OperStatus  int    `json:"oper_status,omitempty"`
	InOctets    int64  `json:"in_octets,omitempty"`
	OutOctets   int64  `json:"out_octets,omitempty"`
}

type IPAddress struct {
	IP      string `json:"ip_address,omitempty"`
	IfIndex int    `json:"interface_index,omitempty"`
}

type PhysicalEntity struct {
	Index        int    `json:"index,omitempty"`
	Description  string `json:"description,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	ModelName    string `json:"model_name,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
}

type CPUMetrics struct {
	User   float64 `json:"user_percent,omitempty"`
	System float64 `json:"system_percent,omitempty"`
	Idle   float64 `json:"idle_percent,omitempty"`
}

type MemoryMetrics struct {
	TotalSwap int64 `json:"total_swap_kb,omitempty"`
	AvailSwap int64 `json:"avail_swap_kb,omitempty"`
}

// OID constants for tables that are not part of the interfaces package.
// Comparisons against walked PDU names use gosnmp's leading-dot string
// format; Get request OIDs use the bare form.
const (
	oidIPAdEntAddr        = ".1.3.6.1.2.1.4.20.1.1"      // ipAdEntAddr (IpAddress)
	oidIPAdEntIfIndex     = ".1.3.6.1.2.1.4.20.1.2"      // ipAdEntIfIndex (Integer)
	oidEntPhysicalDescr   = ".1.3.6.1.2.1.47.1.1.1.1.1"  // entPhysicalDescr
	oidEntPhysicalSerial  = ".1.3.6.1.2.1.47.1.1.1.1.10" // entPhysicalSerialNum
	oidEntPhysicalMfgName = ".1.3.6.1.2.1.47.1.1.1.1.11" // entPhysicalMfgName
	oidEntPhysicalModel   = ".1.3.6.1.2.1.47.1.1.1.1.12" // entPhysicalModelName
	oidSysDescr           = "1.3.6.1.2.1.1.1.0"          // sysDescr
	oidSysObjectID        = "1.3.6.1.2.1.1.2.0"          // sysObjectID
	oidSysUpTime          = "1.3.6.1.2.1.1.3.0"          // sysUpTime (in timeticks)
	oidSysContact         = "1.3.6.1.2.1.1.4.0"          // sysContact
	oidSysName            = "1.3.6.1.2.1.1.5.0"          // sysName
	oidSysLocation        = "1.3.6.1.2.1.1.6.0"          // sysLocation
	oidSSCpuRawUser       = "1.3.6.1.4.1.2021.11.50.0"   // ssCpuRawUser
	oidSSCpuRawSystem     = "1.3.6.1.4.1.2021.11.51.0"   // ssCpuRawSystem
	oidSSCpuRawIdle       = "1.3.6.1.4.1.2021.11.52.0"   // ssCpuRawIdle
	oidMemTotalSwap       = "1.3.6.1.4.1.2021.4.3.0"     // memTotalSwap
	oidMemAvailSwap       = "1.3.6.1.4.1.2021.4.4.0"     // memAvailSwap
)

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
	ifaces, err := collectInterfaces(snmpClient)
	if err != nil {
		log.Printf("Warning: failed to collect interfaces: %v", err)
	} else {
		result.Interfaces = ifaces
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
		oidSysDescr,
		oidSysObjectID,
		oidSysUpTime,
		oidSysContact,
		oidSysName,
		oidSysLocation,
	}

	result, _, err := client.GetValue(context.TODO(), oids)
	if err != nil {
		return nil, err
	}

	for i, oid := range oids {
		if i >= len(result.Variables) {
			break
		}
		value := result.Variables[i].Value
		if oid == oidSysUpTime {
			// Convert timeticks to seconds (1 timetick = 1/100 second);
			// gosnmp decodes TimeTicks to uint32.
			if val, ok := value.(uint32); ok {
				info.UpTime = float64(val) / 100
			}
			continue
		}
		// gosnmp decodes octet strings and OID values (sysObjectID) to
		// []byte and string respectively.
		if val, ok := asOctetString(value); ok {
			switch oid {
			case oidSysDescr:
				info.Description = val
			case oidSysObjectID:
				info.ObjectID = val
			case oidSysContact:
				info.Contact = val
			case oidSysName:
				info.Name = val
			case oidSysLocation:
				info.Location = val
			}
		}
	}

	return info, nil
}

func collectInterfaces(client *snmp.Client) ([]Interface, error) {
	var ifaces []Interface

	// Use Walk to get all interface information
	baseOID := "1.3.6.1.2.1.2.2.1"
	oidsMap, _, err := client.Walk(context.TODO(), baseOID)
	if err != nil {
		return nil, err
	}

	interfaceDetails := make(map[int]*Interface)

	for oid, value := range oidsMap {
		fields := strings.Split(oid, ".")
		if len(fields) < 2 {
			continue
		}

		index := parseInterfaceIndex(fields[len(fields)-1])
		if index == 0 {
			continue
		}

		oidWithoutIndex := strings.Join(fields[:len(fields)-1], ".")

		iface, ok := interfaceDetails[index]
		if !ok {
			iface = &Interface{Index: index}
			interfaceDetails[index] = iface
		}

		applyInterfaceOID(iface, oidWithoutIndex, value)
	}

	for _, iface := range interfaceDetails {
		ifaces = append(ifaces, *iface)
	}

	return ifaces, nil
}

// applyInterfaceOID sets the matching field on iface for an ifTable column
// OID without the interface index. Values use the native shapes gosnmp
// returns: []byte for octet strings, int for Integer objects, and uint for
// Counter32/Gauge32 objects.
func applyInterfaceOID(iface *Interface, oid string, value interface{}) {
	switch interfaces.OID(oid) {
	case interfaces.OIDIfIndex:
		if val, ok := value.(int); ok {
			iface.Index = val
		}
	case interfaces.OIDIfDescr:
		if val, ok := asOctetString(value); ok {
			iface.Description = val
		}
	case interfaces.OIDIfType:
		if val, ok := value.(int); ok {
			iface.Type = val
		}
	case interfaces.OIDIfMTU:
		if val, ok := value.(int); ok {
			iface.MTU = val
		}
	case interfaces.OIDIfSpeed:
		iface.Speed = asCounter(value)
	case interfaces.OIDIfPhysAddress:
		if val, ok := value.([]byte); ok && len(val) == 6 {
			iface.MACAddress = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
				val[0], val[1], val[2], val[3], val[4], val[5])
		}
	case interfaces.OIDIfAdminStatus:
		if val, ok := value.(int); ok {
			iface.AdminStatus = val
		}
	case interfaces.OIDIfOperStatus:
		if val, ok := value.(int); ok {
			iface.OperStatus = val
		}
	case interfaces.OIDIfInOctets:
		iface.InOctets = asCounter(value)
	case interfaces.OIDIfOutOctets:
		iface.OutOctets = asCounter(value)
	}
}

// asCounter converts a Counter32/Gauge32 value (gosnmp uint) or any other
// non-negative integer shape into an int64 counter reading. Returns 0 when
// the value is absent or invalid.
func asCounter(value interface{}) int64 {
	switch v := value.(type) {
	case uint:
		return int64(v)
	case uint64:
		return int64(v)
	case int:
		if v >= 0 {
			return int64(v)
		}
	case int64:
		if v >= 0 {
			return v
		}
	}
	return 0
}

// asOctetString accepts the shapes an SNMP octet string arrives in: gosnmp
// []byte or a plain Go string.
func asOctetString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case []byte:
		return string(v), true
	case string:
		return v, true
	default:
		return "", false
	}
}

// parseInterfaceIndex extracts the interface index from an OID suffix
func parseInterfaceIndex(suffix string) int {
	index := 0
	if _, err := fmt.Sscanf(suffix, "%d", &index); err != nil {
		log.Printf("Warning: failed to parse interface index from %s: %v", suffix, err)
	}
	return index
}

func collectIPAddresses(client *snmp.Client) ([]IPAddress, error) {
	var ipAddresses []IPAddress

	// Use Walk to get IP address table
	baseOID := "1.3.6.1.2.1.4.20.1"
	oidsMap, _, err := client.Walk(context.TODO(), baseOID)
	if err != nil {
		return nil, err
	}

	// First pass: collect ipAdEntIfIndex so each address entry is complete
	// regardless of map iteration order.
	ifIndexByIP := make(map[string]int)
	for oid, value := range oidsMap {
		base, ip, ok := splitIPTableOID(oid)
		if !ok || base != oidIPAdEntIfIndex {
			continue
		}
		// ipAdEntIfIndex is an Integer; gosnmp decodes it to int.
		if val, ok := value.(int); ok {
			ifIndexByIP[ip] = val
		}
	}

	// Second pass: collect ipAdEntAddr entries.
	for oid, value := range oidsMap {
		base, ip, ok := splitIPTableOID(oid)
		if !ok || base != oidIPAdEntAddr {
			continue
		}
		addr, ok := asIPAddress(value)
		if !ok {
			continue
		}
		ipAddresses = append(ipAddresses, IPAddress{IP: addr, IfIndex: ifIndexByIP[ip]})
	}

	return ipAddresses, nil
}

// splitIPTableOID splits an ipAddrTable OID into its column base and the IP
// address forming the table index, e.g.
// ".1.3.6.1.2.1.4.20.1.2.10.0.0.1" → (".1.3.6.1.2.1.4.20.1.2", "10.0.0.1", true).
func splitIPTableOID(oid string) (base, ip string, ok bool) {
	fields := strings.Split(oid, ".")
	if len(fields) < 4 {
		return "", "", false
	}

	ip = strings.Join(fields[len(fields)-4:], ".")
	base = strings.Join(fields[:len(fields)-4], ".")
	return base, ip, true
}

// asIPAddress formats an ipAdEntAddr value; gosnmp decodes IpAddress values
// to a dotted string, with []byte and net.IP accepted defensively.
func asIPAddress(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		if len(v) == 4 {
			return fmt.Sprintf("%d.%d.%d.%d", v[0], v[1], v[2], v[3]), true
		}
	case net.IP:
		return v.String(), true
	}
	return "", false
}

func collectPhysicalEntities(client *snmp.Client) ([]PhysicalEntity, error) {
	var entities []PhysicalEntity

	// Use Walk to get physical entity table
	baseOID := "1.3.6.1.2.1.47.1.1.1.1"
	oidsMap, _, err := client.Walk(context.TODO(), baseOID)
	if err != nil {
		return nil, err
	}

	entityMap := make(map[int]*PhysicalEntity)

	for oid, value := range oidsMap {
		fields := strings.Split(oid, ".")
		if len(fields) < 2 {
			continue
		}

		index := parseInterfaceIndex(fields[len(fields)-1])
		base := strings.Join(fields[:len(fields)-1], ".")

		entity, ok := entityMap[index]
		if !ok {
			entity = &PhysicalEntity{Index: index}
			entityMap[index] = entity
		}

		applyEntityOID(entity, base, value)
	}

	for _, entity := range entityMap {
		entities = append(entities, *entity)
	}

	return entities, nil
}

// applyEntityOID sets the matching field on entity for an entPhysicalTable
// column OID. RFC 6933 column numbering: descr=.1, serialNum=.10, mfgName=.11,
// modelName=.12.
func applyEntityOID(entity *PhysicalEntity, base string, value interface{}) {
	switch base {
	case oidEntPhysicalDescr:
		if val, ok := asOctetString(value); ok {
			entity.Description = val
		}
	case oidEntPhysicalSerial:
		if val, ok := asOctetString(value); ok {
			entity.SerialNumber = val
		}
	case oidEntPhysicalMfgName:
		if val, ok := asOctetString(value); ok {
			entity.Vendor = val
		}
	case oidEntPhysicalModel:
		if val, ok := asOctetString(value); ok {
			entity.ModelName = val
		}
	}
}

func collectCPUMetrics(client *snmp.Client) (*CPUMetrics, error) {
	// Use UCD-SNMP-MIB for CPU metrics (Linux/Unix systems); ssCpuRaw* are
	// Counter32 objects, which gosnmp decodes to uint.
	oids := []string{oidSSCpuRawUser, oidSSCpuRawSystem, oidSSCpuRawIdle}

	result, _, err := client.GetValue(context.TODO(), oids)
	if err != nil {
		return nil, err
	}

	values := make(map[string]float64, len(oids))
	totalTicks := float64(0)
	for i, oid := range oids {
		if i >= len(result.Variables) {
			break
		}
		if val, ok := asCounter32(result.Variables[i].Value); ok {
			values[oid] = val
			totalTicks += val
		}
	}

	if totalTicks > 0 {
		metrics := &CPUMetrics{
			User:   values[oidSSCpuRawUser] / totalTicks * 100,
			System: values[oidSSCpuRawSystem] / totalTicks * 100,
			Idle:   values[oidSSCpuRawIdle] / totalTicks * 100,
		}
		return metrics, nil
	}

	return nil, nil // Not available on this device
}

// asCounter32 accepts uint (gosnmp Counter32/Gauge32) or uint32 values.
func asCounter32(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	default:
		return 0, false
	}
}

func collectMemoryMetrics(client *snmp.Client) (*MemoryMetrics, error) {
	// Use UCD-SNMP-MIB for memory metrics (Linux/Unix systems)
	oids := []string{oidMemTotalSwap, oidMemAvailSwap}

	result, _, err := client.GetValue(context.TODO(), oids)
	if err != nil {
		return nil, err
	}

	metrics := &MemoryMetrics{}
	for i, oid := range oids {
		if i >= len(result.Variables) {
			break
		}
		// memTotalSwap/memAvailSwap are Integer objects; gosnmp decodes
		// them to int.
		if val, ok := result.Variables[i].Value.(int); ok {
			switch oid {
			case oidMemTotalSwap:
				metrics.TotalSwap = int64(val)
			case oidMemAvailSwap:
				metrics.AvailSwap = int64(val)
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
	snmpVersion := &snmp.SNMPVersionFlag{}
	flag.Var(snmpVersion, "snmpVersion", "SNMP protocol version: 2c or 3 (default 2c).")
	v3Username, v3AuthProtocol, v3AuthPassphrase, v3PrivProtocol, v3PrivPassphrase := snmp.AddV3Flags(flag.CommandLine)
	flag.Parse()

	version := snmpVersion.VersionOrDefault()

	snmpClient := snmp.Client{
		Target:    *target,
		Community: *community,
		Version:   version,
	}

	if version == gosnmp.Version3 {
		if err := snmp.ApplyV3Flags(&snmpClient, *v3Username, v3AuthProtocol.Protocol, *v3AuthPassphrase, v3PrivProtocol.Protocol, *v3PrivPassphrase); err != nil {
			log.Fatalf("Invalid SNMP v3 configuration: %v", err)
		}
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
