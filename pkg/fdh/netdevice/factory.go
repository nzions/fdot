package netdevice

import (
	"fmt"
	"strings"

	"github.com/nzions/fdot/pkg/fdh/netdevice/genericaruba"
	"github.com/nzions/fdot/pkg/fdh/netmodel"
	"github.com/nzions/fdot/pkg/fdh/netssh"
)

// DeviceType represents the type of network device
type DeviceType string

// Device type constants
// Generic types (e.g., GenericAruba, GenericCiscoIOS) represent base implementations
// that work across versions. Version-specific types will be added as needed
// (e.g., CiscoNXOS5, CiscoNXOS7 when behavior differs between NX-OS 5.x and 7.x)
const (
	GenericAruba        DeviceType = "generic_aruba"
	GenericCiscoIOS     DeviceType = "generic_cisco_ios"
	GenericCiscoNXOS    DeviceType = "generic_cisco_nxos" // May split into nxos_5, nxos_7, etc.
	GenericJuniperJunOS DeviceType = "generic_juniper_junos"
	GenericAristaEOS    DeviceType = "generic_arista_eos"
	DeviceTypeUnknown   DeviceType = "unknown"
)

// DetectDeviceType performs a quick check of show version output to determine device type
// Returns the device type as a DeviceType constant
func DetectDeviceType(showVersionOutput string) DeviceType {
	output := strings.ToLower(showVersionOutput)

	// Detect HP ProCurve/Aruba
	if strings.Contains(output, "procurve") || strings.Contains(output, "aruba") ||
		strings.Contains(output, "hp j") || strings.Contains(output, "hp k") {
		return GenericAruba
	}

	// Detect Cisco IOS
	if strings.Contains(output, "cisco ios") || strings.Contains(output, "cisco internetwork") {
		return GenericCiscoIOS
	}

	// Detect Cisco NX-OS
	if strings.Contains(output, "cisco nx-os") || strings.Contains(output, "nexus") {
		return GenericCiscoNXOS
	}

	// Detect Juniper JunOS
	if strings.Contains(output, "junos") || strings.Contains(output, "juniper") {
		return GenericJuniperJunOS
	}

	// Detect Arista EOS
	if strings.Contains(output, "arista") || strings.Contains(output, "eos") {
		return GenericAristaEOS
	}

	// Unknown device
	return DeviceTypeUnknown
}

// NewDevice creates a new device based on show version output
// Each device type is responsible for parsing its own show version output
// Returns a Device interface implementation
func NewDevice(sshClient *netssh.Client, showVersionOutput string) (netmodel.Device, error) {
	deviceType := DetectDeviceType(showVersionOutput)

	switch deviceType {
	case GenericAruba:
		device, err := genericaruba.NewDevice(sshClient, showVersionOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to create Aruba device: %w", err)
		}
		return device, nil

	// Add more device types here as they are implemented
	// case GenericCiscoIOS:
	//     device, err := NewCiscoIOSDevice(sshClient, showVersionOutput)
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to create Cisco IOS device: %w", err)
	//     }
	//     return device, nil
	//
	// case GenericCiscoNXOS:
	//     device, err := NewCiscoNXOSDevice(sshClient, showVersionOutput)
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to create Cisco NX-OS device: %w", err)
	//     }
	//     return device, nil

	default:
		return nil, fmt.Errorf("unsupported device type: %s (detected from show version)", deviceType)
	}
}
