package netmodel

import "time"

// CacheConfig holds configuration for command output caching
type CacheConfig struct {
	// Enabled determines if caching is active
	Enabled bool
	// TTL is the time-to-live for cached command outputs
	// If a cached file is older than this duration, it will be refreshed
	TTL time.Duration
	// BaseDir is the base directory for storing cached outputs
	// If empty, a default will be used based on device IP
	BaseDir string
}

// DefaultCacheConfig returns a cache configuration with sensible defaults
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled: true,
		TTL:     5 * time.Minute, // Default 5 minutes
		BaseDir: "",              // Will be set per device
	}
}

// DeviceInfo holds the discovered device information
type DeviceInfo struct {
	// Identification
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	Platform  string `json:"platform"`
	OSVersion string `json:"os_version"`
	Model     string `json:"model"`
	Serial    string `json:"serial"`
	Uptime    string `json:"uptime"`

	// Discovery metadata
	DiscoveredAt time.Time `json:"discovered_at"`
	LastUpdated  time.Time `json:"last_updated"`

	// Configuration and operational data
	Interfaces []Interface `json:"interfaces"`
	Neighbors  []Neighbor  `json:"neighbors"`

	// Raw command outputs (for reference)
	RawOutputDir string `json:"raw_output_dir"`
}

// Interface represents a network interface on the device
type Interface struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IPAddress   string `json:"ip_address"`
	Subnet      string `json:"subnet"`
	VRF         string `json:"vrf,omitempty"` // VRF (Virtual Routing and Forwarding) instance
	Status      string `json:"status"`        // up/down
	Protocol    string `json:"protocol"`      // up/down
	VLANs       []int  `json:"vlans"`
}

// Neighbor represents a discovered neighbor via LLDP/CDP
type Neighbor struct {
	LocalInterface  string `json:"local_interface"`
	RemoteHostname  string `json:"remote_hostname"`
	RemoteInterface string `json:"remote_interface"`
	Platform        string `json:"platform"`
	IPAddress       string `json:"ip_address"`
	Capabilities    string `json:"capabilities"`
}

// CommandOutput stores raw command output for a device
type CommandOutput struct {
	DeviceIP   string    `json:"device_ip"`
	Command    string    `json:"command"`
	Output     string    `json:"output"`
	ExecutedAt time.Time `json:"executed_at"`
	Error      string    `json:"error,omitempty"`
}
