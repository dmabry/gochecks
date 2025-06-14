





# check_interface_usage

`check_interface_usage` is a monitoring check that monitors interface usage statistics and utilization from network devices using SNMP.

## Usage

```bash
./cmd/check_interface_usage/check_interface_usage -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")

## Metrics Collected

`check_interface_usage` focuses on interface utilization metrics, including:

- Traffic counters
- Utilization percentages
- Error and discard rates


