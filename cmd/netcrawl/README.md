# netcrawl - Network Device Discovery Tool

A tool for discovering and inventorying HP ProCurve/Aruba network switches via SSH.

## Features

- **HP/Aruba switch support**: Optimized for HP ProCurve, Aruba switches, and Comware devices
- **SSH connectivity**: Uses optimized SSH settings for network devices
- **Command execution**: Executes standard show commands:
  - `show version` - Platform, OS version, model, serial number
  - `show running-config` - Interface/VLAN IP addresses, VRF assignments
  - `show lldp neighbors detail` - Neighbor discovery
- **Data persistence**: 
  - Saves raw command outputs to text files
  - Parses and stores structured data in JSON database (dsjdb)
- **Interface parsing**: Captures IP addresses, descriptions, VRF instances, VLAN assignments
- **Error handling**: Captures and reports errors for each command

## Building

```bash
go build -o bin/netcrawl ./cmd/netcrawl
```

## Prerequisites

1. **Set SSH credentials** (required before first run):
   ```bash
   credmgr setssh <username> <password>
   ```

2. **Set encryption key** (Linux only):
   ```bash
   # Generate a random key
   openssl rand -hex 32
   
   # Set it in your environment
   export CREDMGR_KEY="<your-64-hex-character-key>"
   
   # Make it persistent (add to ~/.bashrc)
   echo 'export CREDMGR_KEY="<your-key>"' >> ~/.bashrc
   ```

## Usage

```bash
# Crawl a single device (required flag)
./bin/netcrawl -device <ip-address>

# Example
./bin/netcrawl -device 192.168.1.1

# With custom port
./bin/netcrawl -device 192.168.1.1 -port 2222

# With custom timeout
./bin/netcrawl -device 192.168.1.1 -timeout 60s

# Show all options
./bin/netcrawl -h
```

### Command-Line Flags

- `-device` (string, **required**): Target device IP address
- `-port` (int, default: 22): SSH port number
- `-timeout` (duration, default: 30s): Connection timeout (e.g., 30s, 1m, 90s)

## Output

### Text Files

Raw command outputs are saved to:
```
~/.fdot/netcfg/<ip-address>/
├── show_version.txt
├── show_running_config.txt
└── show_lldp_neighbors.txt
```

### JSON Database

Parsed device data is saved to:
```
~/.fdot/devices/<ip-address>.json
```

The JSON structure includes:
```json
{
  "hostname": "hp-switch01",
  "ip_address": "192.168.1.1",
  "platform": "ProCurve",
  "os_version": "WB.16.11.0012",
  "model": "HP J9624A 2620-48",
  "serial": "SG35FCX8J2",
  "uptime": "5 days",
  "discovered_at": "2025-10-17T10:30:00Z",
  "last_updated": "2025-10-17T10:30:00Z",
  "interfaces": [
    {
      "name": "1",
      "description": "Uplink to Core",
      "ip_address": "",
      "subnet": "",
      "vlans": [10, 20, 30]
    },
    {
      "name": "vlan10",
      "description": "Management VLAN",
      "ip_address": "10.1.1.1",
      "subnet": "255.255.255.0",
      "vrf": "MGMT"
    }
  ],
  "neighbors": [
    {
      "local_interface": "24",
      "remote_hostname": "aruba-core",
      "remote_interface": "1/1/1",
      "ip_address": "10.1.1.254",
      "platform": "Aruba 2930F Switch",
      "capabilities": "Bridge, Router"
    }
  ],
  "raw_output_dir": "/home/user/.fdot/netcfg/192.168.1.1"
}
```

## Example Run

```
$ ./bin/netcrawl -device 192.168.1.1

NetCrawl - Network Device Discovery Tool
=========================================
✓ Loaded SSH credentials for user: admin
✓ Target device: 192.168.1.1:22
→ Connecting to device...
✓ Connected successfully
→ Executing: show version
✓ Saved output to: /home/user/.fdot/netcfg/192.168.1.1/show_version.txt
→ Executing: show running-config
✓ Saved output to: /home/user/.fdot/netcfg/192.168.1.1/show_running_config.txt
→ Executing: show lldp neighbors detail
✓ Saved output to: /home/user/.fdot/netcfg/192.168.1.1/show_lldp_neighbors.txt
→ Parsing device information...
✓ Parsed device info:
  Platform:  ProCurve
  OS:        WB.16.11.0012
  Model:     HP J9624A 2620-48
  Serial:    SG35FCX8J2
  Uptime:    5 days
  Interfaces: 48
  Neighbors:  3
→ Saving to database...
✓ Device data saved to database: /home/user/.fdot/devices/192.168.1.1.json

✅ NetCrawl completed successfully!
```

## Error Handling

If a command fails to execute, the error is reported and the tool continues:

```
→ Executing: show lldp neighbors detail
⚠ Warning: command failed: command execution timeout
```

## Supported Devices

Currently optimized for **HP ProCurve/Aruba switches**:
- HP ProCurve switches (J-series, K-series)
- Aruba switches (2530, 2540, 2920, 2930, 3810, 5400, etc.)
- HP Comware devices (A-series, H3C)

The parser handles:
- HP ProCurve syntax for interfaces and VLANs
- Aruba ArubaOS-Switch configuration format
- LLDP neighbor discovery (HP/Aruba format)
- Both legacy and modern IP configuration syntax
- VLAN tagging (tagged/untagged ports)

Future support could include:
- Additional HP/Aruba models
- Cisco IOS (modify parser patterns)
- Arista EOS
- Juniper JunOS
- Contributions welcome!

## Architecture

- **pkg/fdh/netssh**: SSH client wrapper with network device optimizations
- **pkg/fdh/netdevice**: Data structures and parser for device information
- **cmd/netcrawl**: Main application

## Troubleshooting

### "no ssh credentials found"
Set credentials using: `credmgr setssh <username> <password>`

### "CREDMGR_KEY environment variable not set" (Linux)
Generate and set the encryption key (see Prerequisites above)

### "failed to connect"
- Verify device IP is reachable: `ping <ip-address>`
- Check SSH is enabled on device: `ssh <username>@<ip-address>`
- Verify credentials are correct
- Check firewall rules

### "command failed"
Some devices may not support all commands. The tool will continue with other commands.

## Development

### Adding Support for New Vendors

1. Update parser in `pkg/fdh/netdevice/parser.go`
2. Add vendor-specific regex patterns
3. Test with sample output files
4. Submit PR with test cases

### Modifying Commands

Edit the `commands` map in `cmd/netcrawl/netcrawl.go`:

```go
commands := map[string]string{
    "show version": "show_version.txt",
    "show inventory": "show_inventory.txt",  // Add new command
}
```

## Version

1.0.0 - Initial release

## License

Property of FDOT (Florida Department of Transport)
