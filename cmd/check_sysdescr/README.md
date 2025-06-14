





# check_sysdescr

`check_sysdescr` is a monitoring check that retrieves and reports system description information from network devices using SNMP.

## Usage

```bash
./cmd/check_sysdescr/check_sysdescr -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")

## Metrics Collected

`check_sysdescr` collects system description information, including:

- Device model
- Software version
- Hardware serial number



