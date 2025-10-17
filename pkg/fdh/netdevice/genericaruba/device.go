package genericaruba

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nzions/fdot/pkg/fdh/netmodel"
	"github.com/nzions/fdot/pkg/fdh/netssh"
)

// Compile-time check to ensure Device implements netmodel.Device interface
var _ netmodel.Device = (*Device)(nil)

// Device represents HP ProCurve and Aruba switches (ArubaOS-Switch, version 10.x style)
type Device struct {
	client *netssh.Client
	info   *netmodel.DeviceInfo
}

// NewDevice creates a new Aruba device instance by parsing show version output
func NewDevice(client *netssh.Client, showVersionOutput string) (*Device, error) {
	// Parse show version to extract device information
	platform, osVersion, model, serial, uptime := ParseShowVersion(showVersionOutput)

	if platform == "" && model == "" {
		return nil, fmt.Errorf("failed to parse Aruba device information from show version")
	}

	return &Device{
		client: client,
		info: &netmodel.DeviceInfo{
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

// GetHostname returns the device hostname
func (d *Device) GetHostname() string {
	return d.info.Hostname
}

// GetIPAddress returns the device IP address
func (d *Device) GetIPAddress() string {
	return d.info.IPAddress
}

// GetPlatform returns the device platform
func (d *Device) GetPlatform() string {
	return d.info.Platform
}

// GetOSVersion returns the device OS version
func (d *Device) GetOSVersion() string {
	return d.info.OSVersion
}

// GetModel returns the device model
func (d *Device) GetModel() string {
	return d.info.Model
}

// GetSerial returns the device serial number
func (d *Device) GetSerial() string {
	return d.info.Serial
}

// GetUptime returns the device uptime
func (d *Device) GetUptime() string {
	return d.info.Uptime
}

// GetConfig retrieves and parses the running configuration
func (d *Device) GetConfig() (string, error) {
	if !d.IsConnected() {
		return "", fmt.Errorf("device not connected")
	}
	return d.client.ExecuteCommand("show running-config")
}

// GetInterfaces retrieves and parses interface information
func (d *Device) GetInterfaces() ([]netmodel.Interface, error) {
	if !d.IsConnected() {
		return nil, fmt.Errorf("device not connected")
	}

	config, err := d.GetConfig()
	if err != nil {
		return nil, err
	}

	interfaces := d.parseInterfaces(config)
	d.info.Interfaces = interfaces
	d.info.LastUpdated = time.Now()

	return interfaces, nil
}

// GetNeighbors retrieves and parses LLDP neighbor information
func (d *Device) GetNeighbors() ([]netmodel.Neighbor, error) {
	if !d.IsConnected() {
		return nil, fmt.Errorf("device not connected")
	}

	output, err := d.client.ExecuteCommand("show lldp neighbors detail")
	if err != nil {
		return nil, err
	}

	neighbors := d.parseNeighbors(output)
	d.info.Neighbors = neighbors
	d.info.LastUpdated = time.Now()

	return neighbors, nil
}

// GetDeviceInfo returns the device information structure
func (d *Device) GetDeviceInfo() *netmodel.DeviceInfo {
	return d.info
}

// SetIPAddress sets the device IP address
func (d *Device) SetIPAddress(ip string) {
	d.info.IPAddress = ip
}

// Connect establishes SSH connection (if not already connected)
func (d *Device) Connect() error {
	if d.client == nil {
		return fmt.Errorf("no SSH client configured")
	}
	// The client is already connected by the time device is created
	// This is a no-op but satisfies the interface
	return nil
}

// Disconnect closes the SSH connection
func (d *Device) Disconnect() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// IsConnected checks if the device is connected
func (d *Device) IsConnected() bool {
	return d.client != nil
}

// parseInterfaces parses HP/Aruba running-config for interface information
func (d *Device) parseInterfaces(config string) []netmodel.Interface {
	var interfaces []netmodel.Interface
	var currentInterface *netmodel.Interface

	scanner := bufio.NewScanner(strings.NewReader(config))

	// HP switches use different interface naming
	interfaceRe := regexp.MustCompile(`^(?:interface|vlan)\s+([\w/.-]+)`)
	ipRe := regexp.MustCompile(`^\s+ip address\s+([\d.]+)(?:\s+([\d.]+)|/([\d]+))`)
	descRe := regexp.MustCompile(`^\s+(?:name|description)\s+(.+)`)
	vrfRe := regexp.MustCompile(`^\s+(?:vrf attach|ip vrf forwarding|vrf forwarding)\s+([\w-]+)`)
	vlanRe := regexp.MustCompile(`^\s+(?:tagged|untagged)\s+vlan\s+([\d,\-]+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Start of new interface/VLAN block
		if match := interfaceRe.FindStringSubmatch(line); match != nil {
			// Save previous interface if exists
			if currentInterface != nil {
				interfaces = append(interfaces, *currentInterface)
			}
			// Start new interface
			currentInterface = &netmodel.Interface{
				Name: match[1],
			}
			continue
		}

		if currentInterface == nil {
			continue
		}

		// Parse IP address (supports both netmask and CIDR notation)
		if match := ipRe.FindStringSubmatch(line); match != nil {
			currentInterface.IPAddress = match[1]
			if match[2] != "" {
				currentInterface.Subnet = match[2]
			} else if match[3] != "" {
				currentInterface.Subnet = "/" + match[3]
			}
		}

		// Parse name/description
		if match := descRe.FindStringSubmatch(line); match != nil {
			currentInterface.Description = strings.TrimSpace(match[1])
		}

		// Parse VRF
		if match := vrfRe.FindStringSubmatch(line); match != nil {
			currentInterface.VRF = strings.TrimSpace(match[1])
		}

		// Parse VLAN assignments
		if match := vlanRe.FindStringSubmatch(line); match != nil {
			vlanStr := match[1]
			vlans := parseVLANString(vlanStr)
			currentInterface.VLANs = append(currentInterface.VLANs, vlans...)
		}

		// HP uses "exit" to end blocks
		if strings.TrimSpace(line) == "exit" {
			interfaces = append(interfaces, *currentInterface)
			currentInterface = nil
		}
	}

	// Add last interface if exists
	if currentInterface != nil {
		interfaces = append(interfaces, *currentInterface)
	}

	return interfaces
}

// parseNeighbors parses HP/Aruba LLDP neighbor output
func (d *Device) parseNeighbors(output string) []netmodel.Neighbor {
	var neighbors []netmodel.Neighbor
	var currentNeighbor *netmodel.Neighbor

	scanner := bufio.NewScanner(strings.NewReader(output))

	localPortRe := regexp.MustCompile(`(?i)(?:LocalPort|Local Port|^\s*\|\s*)\s*[:=]?\s*(\d+|[\w/.-]+)`)
	tableEntryRe := regexp.MustCompile(`^\s*\|\s*(\d+)\s*\|\s*([\w:.-]+)\s+(.+?)(?:\s*\||\s*$)`)
	remoteIntfRe := regexp.MustCompile(`(?i)(?:Port Descr|ChassisId|Port\s*ID)\s*:?\s*(.+)`)
	hostnameRe := regexp.MustCompile(`(?i)(?:System Name|SysName)\s*:?\s*(.+)`)
	capabilitiesRe := regexp.MustCompile(`(?i)(?:System Capabilities|Capabilities)\s*:?\s*(.+)`)
	ipRe := regexp.MustCompile(`(?i)(?:Management Address|Mgmt Address|Address)\s*:?\s*([\d.]+)`)
	platformRe := regexp.MustCompile(`(?i)(?:System Descr|System Description)\s*:?\s*(.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "===") {
			continue
		}

		// Check for table entry format
		if match := tableEntryRe.FindStringSubmatch(line); match != nil {
			if currentNeighbor != nil {
				neighbors = append(neighbors, *currentNeighbor)
			}
			currentNeighbor = &netmodel.Neighbor{
				LocalInterface: match[1],
			}
			if match[3] != "" {
				parts := strings.Fields(match[3])
				if len(parts) > 0 {
					currentNeighbor.RemoteHostname = parts[0]
				}
			}
			continue
		}

		// Local port indicates start of new neighbor entry
		if match := localPortRe.FindStringSubmatch(line); match != nil {
			if currentNeighbor != nil {
				neighbors = append(neighbors, *currentNeighbor)
			}
			currentNeighbor = &netmodel.Neighbor{
				LocalInterface: match[1],
			}
			continue
		}

		if currentNeighbor == nil {
			continue
		}

		if match := remoteIntfRe.FindStringSubmatch(line); match != nil {
			currentNeighbor.RemoteInterface = strings.TrimSpace(match[1])
		}

		if match := hostnameRe.FindStringSubmatch(line); match != nil {
			currentNeighbor.RemoteHostname = strings.TrimSpace(match[1])
		}

		if match := capabilitiesRe.FindStringSubmatch(line); match != nil {
			currentNeighbor.Capabilities = strings.TrimSpace(match[1])
		}

		if match := ipRe.FindStringSubmatch(line); match != nil {
			currentNeighbor.IPAddress = match[1]
		}

		if match := platformRe.FindStringSubmatch(line); match != nil {
			currentNeighbor.Platform = strings.TrimSpace(match[1])
		}
	}

	// Add last neighbor
	if currentNeighbor != nil {
		neighbors = append(neighbors, *currentNeighbor)
	}

	return neighbors
}

// parseVLANString parses a VLAN string like "1,5,10-15" into a slice of VLAN IDs
func parseVLANString(vlanStr string) []int {
	var vlans []int

	parts := strings.Split(vlanStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err1 == nil && err2 == nil {
					for i := start; i <= end; i++ {
						vlans = append(vlans, i)
					}
				}
			}
		} else {
			vlanID, err := strconv.Atoi(part)
			if err == nil {
				vlans = append(vlans, vlanID)
			}
		}
	}

	return vlans
}
