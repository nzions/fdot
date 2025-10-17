package netmodel

// Device is the interface that all network devices must implement
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
