# check_bgp_peers

`check_bgp_peers` is a monitoring check that monitors BGP peer relationships on network devices using SNMP and the BGP4-MIB (RFC 4273).

## Usage

```bash
./cmd/check_bgp_peers/check_bgp_peers -target 192.168.1.1 -community public
```

### Options

- `-target`: The IP address or hostname of the SNMP target device (default: "127.0.0.1")
- `-community`: The SNMP community string (default: "public")
- `-enablePerfData`: Enable performance data output (default: false)
- `-snmpVersion`: SNMP protocol version: 2c or 3 (default 2c)
- SNMP v3 flags: see the [SNMP v3](../README.md#snmp-v3) section in the main README

## Behavior

The check walks the `bgpPeerIdentifier` column (`1.3.6.1.2.1.15.4.1.1`) to enumerate peers, then queries `bgpPeerAdminStatus` (`1.3.6.1.2.1.15.4.1.8`) and `bgpPeerState` (`1.3.6.1.2.1.15.4.1.9`) for each peer. Peers that cannot be queried are skipped.

A peer is reported as mismatched when its administrative status and session state are inconsistent:

- Administratively enabled (`start`) but the session is not `established` — the BGP session is down or stuck in a transitional state
- Administratively stopped (`stop`) while the session is `established`

## Results

- **OK**: All peers have consistent admin status and session state (including a device with no BGP peers)
- **Critical**: Any peer has an admin status mismatch, or the BGP peer table could not be retrieved

## Metrics Collected

With `-enablePerfData`:

- `total_peers`: Total number of BGP peers discovered
- `mismatched_peers`: Number of peers with an admin status mismatch
