# NetDevice Package Architecture

## Overview

The `netdevice` package uses an interface-based design pattern to support multiple network device vendors and operating systems. This architecture allows for easy extensibility and vendor-specific implementations.

## Design Pattern

### Device Interface

All network devices must implement the `Device` interface defined in `types.go`:

```go
type Device interface {
    // Discovery and identification
    GetHostname() string
    GetIPAddress() string
    GetPlatform() string
    GetOSVersion() string
    GetModel() string
    GetSerial() string
    GetUptime() string

    // Configuration operations
    GetConfig() (string, error)
    GetInterfaces() ([]Interface, error)
    GetNeighbors() ([]Neighbor, error)

    // Data access
    GetDeviceInfo() *DeviceInfo
    SetIPAddress(ip string)

    // Lifecycle
    Connect() error
    Disconnect() error
    IsConnected() bool
}
```

### Factory Pattern

The `factory.go` file provides:

1. **DeviceType** - Type-safe constants for device types:
   - `DeviceTypeAruba` - HP ProCurve/Aruba switches
   - `DeviceTypeCiscoIOS` - Cisco IOS devices
   - `DeviceTypeCiscoNXOS` - Cisco NX-OS devices
   - `DeviceTypeJuniperJunOS` - Juniper JunOS devices
   - `DeviceTypeAristaEOS` - Arista EOS devices
   - `DeviceTypeUnknown` - Unknown/unsupported devices

2. **DetectDeviceType()** - Quick check of `show version` output to identify device type (returns DeviceType constant)

3. **NewDevice()** - Creates the appropriate device implementation, which parses its own show version output

**Key principle**: Each device is responsible for parsing its own show version output.

### Device Creation Flow

```
1. Execute "show version" on device
2. factory.DetectDeviceType() → Quick string matching → Returns device type
3. factory.NewDevice() → Calls device-specific constructor
4. Device constructor (e.g., NewArubaDevice) → Parses show version → Returns device instance
```

## Adding a New Device Type

To add support for a new device vendor or OS:

### 1. Create Device File

Create a new file named `<vendor>_<os_version>.go` (e.g., `cisco_ios.go`, `aruba_10.go`, `juniper_junos.go`):

```go
package netdevice

import (
    "fmt"
    "regexp"
    "strings"
    "time"
    "github.com/nzions/fdot/pkg/fdh/netssh"
)

// MyVendorDevice represents <vendor> <model/os> devices
type MyVendorDevice struct {
    client *netssh.Client
    info   *DeviceInfo
}

// NewMyVendorDevice creates a new device instance
// It parses the show version output internally
func NewMyVendorDevice(client *netssh.Client, showVersionOutput string) (*MyVendorDevice, error) {
    // Parse show version to extract device information
    platform, osVersion, model, serial, uptime := parseMyVendorShowVersion(showVersionOutput)

    if platform == "" {
        return nil, fmt.Errorf("failed to parse device information from show version")
    }

    return &MyVendorDevice{
        client: client,
        info: &DeviceInfo{
            Platform:     platform,
            OSVersion:    osVersion,
            Model:        model,
            Serial:       serial,
            Uptime:       uptime,
            DiscoveredAt: time.Now(),
            LastUpdated:  time.Now(),
        },
    }, nil
}

// parseMyVendorShowVersion parses vendor-specific show version output
func parseMyVendorShowVersion(output string) (platform, osVersion, model, serial, uptime string) {
    // Vendor-specific parsing logic using regex
    platformRe := regexp.MustCompile(`pattern for platform`)
    // ... extract fields
    return
}

// Implement all Device interface methods...
// GetHostname(), GetIPAddress(), GetPlatform(), etc.

// Compile-time check to ensure device implements Device interface
var _ Device = (*MyVendorDevice)(nil)
```

### 2. Add Detection Logic

Update `factory.go` to detect your device type:

```go
// Add your device type constant
const (
    DeviceTypeMyVendor DeviceType = "myvendor"
    // ... existing constants
)

func DetectDeviceType(showVersionOutput string) DeviceType {
    output := strings.ToLower(showVersionOutput)

    // Add detection for your vendor (quick check only)
    if strings.Contains(output, "myvendor") {
        return DeviceTypeMyVendor
    }

    // Existing detection...
}

func NewDevice(sshClient *netssh.Client, showVersionOutput string) (Device, error) {
    deviceType := DetectDeviceType(showVersionOutput)

    switch deviceType {
    case DeviceTypeMyVendor:
        device, err := NewMyVendorDevice(sshClient, showVersionOutput)
        if err != nil {
            return nil, fmt.Errorf("failed to create MyVendor device: %w", err)
        }
        return device, nil

    // Existing cases...
    }
}
```

**Important**: 
- Use type-safe DeviceType constants instead of strings
- `DetectDeviceType()` should only do quick string matching
- Full parsing happens in the device constructor

### 3. Implement Parsing Logic

All parsing logic stays in your device file:

```go
// parseMyVendorShowVersion parses vendor-specific show version output
// This is called by the constructor
func parseMyVendorShowVersion(output string) (platform, osVersion, model, serial, uptime string) {
    // Vendor-specific regex patterns
    platformRe := regexp.MustCompile(`your pattern`)
    // ... parse all fields
    return
}

// parseInterfaces parses vendor-specific config format
func (d *MyVendorDevice) parseInterfaces(config string) []Interface {
    // Vendor-specific parsing logic
}

// parseNeighbors parses vendor-specific neighbor output
func (d *MyVendorDevice) parseNeighbors(output string) []Neighbor {
    // Vendor-specific parsing logic
}
```

### 4. Interface Compliance Check

**CRITICAL**: Add this line at the bottom of your device file:

```go
// Compile-time check to ensure MyVendorDevice implements Device interface
var _ Device = (*MyVendorDevice)(nil)
```

This ensures your device type implements all required interface methods. If you miss any methods, the compiler will fail with a clear error message.

## Current Implementations

### aruba_10.go

Supports:
- HP ProCurve switches (J-series, K-series)
- Aruba ArubaOS-Switch (2530, 2540, 2920, 2930, 3810, 5400, etc.)
- HP Comware devices

Features:
- Parses HP-specific interface naming (`interface 1`, `vlan 10`)
- Handles VLAN tagging (tagged/untagged syntax)
- Supports VRF parsing
- LLDP neighbor discovery
- Configuration backup

Interface compliance: ✅ `var _ Device = (*ArubaDevice)(nil)`

## File Naming Convention

Device files should follow this naming pattern:

- `<vendor>_<os_major_version>.go` - For version-specific implementations
- Examples:
  - `aruba_10.go` - Aruba ArubaOS-Switch 10.x
  - `cisco_ios.go` - Cisco IOS (all versions if behavior is consistent)
  - `cisco_iosxe.go` - Cisco IOS-XE (if different from IOS)
  - `cisco_nxos.go` - Cisco NX-OS
  - `juniper_junos.go` - Juniper JunOS
  - `arista_eos.go` - Arista EOS

## Testing

When adding a new device type, verify:

1. **Compilation**: `go build ./cmd/netcrawl`
2. **Interface compliance**: The `var _ Device = (*YourDevice)(nil)` line will fail if methods are missing
3. **Detection**: Run against actual device to verify detection works
4. **Parsing**: Verify all fields are correctly parsed
5. **Storage**: Check JSON output has complete data

## Common Pitfalls

1. **Forgetting methods**: Always add the interface compliance check
2. **Package naming**: All device files must be in `package netdevice`
3. **Regex patterns**: Test regex with actual device output, not documentation examples
4. **Error handling**: Return errors, don't panic
5. **Memory leaks**: Always close SSH connections in Disconnect()

## Benefits of This Architecture

1. **Type Safety**: Compiler ensures all devices implement required methods
2. **Extensibility**: Easy to add new vendors without modifying existing code
3. **Testability**: Each device type can be tested independently
4. **Maintainability**: Vendor-specific logic is isolated in separate files
5. **Discovery**: Automatic device type detection from show version output
