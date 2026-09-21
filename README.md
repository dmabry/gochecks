

# gochecks

gochecks is a collection of monitoring checks designed to work with Nagios/Icinga platforms. These checks use SNMP (Simple Network Management Protocol) to gather information from network devices and report their status.

## Features

- Multiple check types for different aspects of network device monitoring
- Cross-platform binaries for Linux, Windows, and macOS
- Package formats: RPM, DEB, APK
- Lightweight and efficient using Go language
- Compatible with gomonitor for result reporting

## Available Checks

gochecks currently includes the following checks:

1. **check_interfaces**: Monitors interface metrics such as status, speed, traffic counters, etc.
2. **check_interface_usage**: Monitors interface usage statistics and utilization
3. **check_sysdescr**: Checks system description information from network devices
4. **device_inventory**: SNMP-based inventory collection (system info, interfaces, IP addresses, hardware)

## Installation

### From Source

To build gochecks from source, you'll need Go installed on your system.

```bash
# Clone the repository
git clone https://github.com/dmabry/gochecks.git

# Change to the project directory
cd gochecks

# Build multi-platform binaries
./scripts/build-multiplatform.sh

# Build packages (RPM, DEB, APK)
./scripts/build-packages.sh
```

### Pre-built Binaries and Packages

Pre-built binaries and packages are available in the `bin` and `build/package/release` directories respectively.

## Usage

Each check has its own set of command-line options. Here are some examples:

**Check all interfaces:**

```bash
./cmd/check_interfaces/check_interfaces -target 192.168.1.1 -community public
```

**Filter by interface description pattern:**

```bash
./cmd/check_interfaces/check_interfaces -target 192.168.1.1 -community public -iface "GigabitEthernet"
```

### Common Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")

### check_interfaces Specific Options

- `-iface`: Filter interfaces by description pattern (optional). Only shows interfaces whose description matches this pattern.

### check_interface_usage

Monitors interface usage by taking two measurements separated by a delay and computing traffic rates. Thresholds (`-warnIn`, `-critIn`, `-warnOut`, `-critOut`) are specified in bps.

```bash
./cmd/check_interface_usage/check_interface_usage -target 192.168.1.1 -community public -index 2 -delay 10 -warnIn 500000000 -critIn 900000000 -enablePerfData
```

### check_sysdescr

Checks the device system description (sysDescr) against an optional regex pattern.

```bash
./cmd/check_sysdescr/check_sysdescr -target 192.168.1.1 -community public -sysDescrPattern "Cisco IOS"
```

### device_inventory

Comprehensive SNMP-based inventory collection: system information, interfaces, IP addresses, physical hardware, and optional CPU/memory metrics. Outputs JSON. See [cmd/device_inventory/README.md](cmd/device_inventory/README.md) for details.

```bash
./cmd/device_inventory/device_inventory -target 192.168.1.1 -community public
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

gochecks is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
