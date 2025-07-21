
# device_inventory

`device_inventory` is a comprehensive SNMP-based inventory tool that collects detailed information about network devices using standard MIBs.

## Usage

```bash
./cmd/device_inventory/device_inventory -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")
- `-output`: Output format (currently only "json" is supported)

## Information Collected

The tool collects a comprehensive set of inventory information, including:

### System Information
- Device description
- Object ID (vendor identifier)
- System uptime
- Contact information
- System name and location

### Interface Information
- Interface index, description, and type
- MTU size and speed
- MAC address
- Administrative and operational status
- Traffic statistics (bytes in/out)

### IP Address Information
- All configured IP addresses and their associated interfaces

### Hardware Inventory
- Physical entity descriptions
- Vendor names
- Model names
- Serial numbers

### Performance Metrics (optional)
- CPU usage percentages (user, system, idle)
- Memory metrics (total and available swap)

## Output Format

The tool outputs the collected information in JSON format by default. Example output:

```json
{
  "system_info": {
    "description": "Cisco IOS Software, C1900 Software (C1900-UNIVERSALK9-M), Version 15.4(3)M6, RELEASE SOFTWARE (fc2)",
    "object_id": "1.3.6.1.4.1.9.1.1018",
    "uptime_seconds": 1234567.89,
    "contact": "admin@example.com",
    "name": "router1",
    "location": "Data Center Rack A"
  },
  "interfaces": [
    {
      "index": 1,
      "description": "GigabitEthernet0/0",
      "type": 62, // ethernetCsmacd (62)
      "mtu": 1500,
      "speed_bps": 1000000000,
      "mac_address": "00:1A:2B:3C:4D:5E",
      "admin_status": 1, // up (1)
      "oper_status": 1, // up (1)
      "in_octets": 1234567890,
      "out_octets": 987654321
    }
  ],
  "ip_addresses": [
    {
      "ip_address": "192.168.1.1",
      "interface_index": 1
    }
  ],
  "physical_entities": [
    {
      "index": 1,
      "description": "Chassis",
      "vendor": "Cisco",
      "model_name": "C1900",
      "serial_number": "FTX12345678"
    }
  ],
  "cpu": {
    "user_percent": 15.2,
    "system_percent": 10.3,
    "idle_percent": 74.5
  },
  "memory": {
    "total_swap_kb": 2097152,
    "avail_swap_kb": 1897152
  }
}
```

## Dependencies

This tool requires the `gochecks` internal SNMP package.

## Building

To build the tool, use:

```bash
go build -o device_inventory ./cmd/device_inventory/device_inventory.go
```
