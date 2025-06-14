




# check_interfaces

`check_interfaces` is a monitoring check that retrieves and reports interface metrics from network devices using SNMP.

## Usage

```bash
./cmd/check_interfaces/check_interfaces -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")

## Metrics Collected

`check_interfaces` collects a wide range of interface metrics, including:

- Interface description, name, and alias
- MAC address
- Interface type and MTU
- Speed and high speed
- Administrative and operational status
- Traffic counters (octets, packets)
- Error and discard counters
- And more...



