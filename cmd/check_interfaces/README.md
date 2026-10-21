




# check_interfaces

`check_interfaces` is a monitoring check that retrieves and reports interface metrics from network devices using SNMP.

## Usage

```bash
./cmd/check_interfaces/check_interfaces -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")
- `-iface`: Filter interfaces by description prefix (optional). Only shows interfaces whose description starts with this value.
- `-enablePerfData`: Add an `interfaces` performance metric counting the reported interfaces (after filtering).

## Filtering

`-iface` filters the rendered interface list by description prefix. When no interface matches, the check returns `Critical`. The `interfaces` performance metric, when enabled, counts only the interfaces included in the result.

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



