






# check_bgp_peers

`check_bgp_peers` is a monitoring check that monitors BGP peer relationships on network devices using SNMP.

## Usage

```bash
./cmd/check_bgp_peers/check_bgp_peers -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")

## Metrics Collected

`check_bgp_peers` collects BGP peer relationship information, including:

- Peer IP addresses
- Peer AS numbers
- Peer status




