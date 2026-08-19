
# check_interface_usage

`check_interface_usage` is a monitoring check that monitors interface usage statistics and utilization from network devices using SNMP. It takes two measurements separated by a delay and reports the inbound/outbound rate against warning and critical thresholds.

## Usage

```bash
./cmd/check_interface_usage/check_interface_usage -target 192.168.1.1 -community public -index 1
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")
- `-index`: The index of the interface to monitor (default: 1)
- `-delay`: The delay in seconds to wait between the two measurements (default: 10)
- `-enablePerfData`: Enable performance data output (default: false)
- `-warnIn`: Warning threshold for inbound traffic in bps (default: 0)
- `-critIn`: Critical threshold for inbound traffic in bps (default: 0)
- `-warnOut`: Warning threshold for outbound traffic in bps (default: 0)
- `-critOut`: Critical threshold for outbound traffic in bps (default: 0)
- `-checkStatus`: Check the interface admin/oper status before measuring and return Critical if the interface is down (default: true). Set to `false` to measure regardless of interface status.

## Interface Status Check

When `-checkStatus` is enabled (the default), the check first queries the interface's `ifAdminStatus` and `ifOperStatus` OIDs. If the interface is not up (either status is not `1`), the check returns `Critical` without measuring. This avoids reporting misleading utilization metrics for interfaces that are administratively down or otherwise not operational.

## Metrics Collected

`check_interface_usage` focuses on interface utilization metrics, including:

- Inbound and outbound traffic rates (32-bit and high-capacity 64-bit counters)
- Utilization compared against warning and critical thresholds
- SNMP request latency (when `-enablePerfData` is set)
