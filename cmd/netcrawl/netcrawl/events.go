package netcrawl

import "time"

type DiscoveryStarted struct {
	IP       string
	Port     int
	Username string
}

type ShowVersionRetrieved struct {
	IP           string
	OutputLength int
	SavedTo      string
}

type DeviceDetected struct {
	IP       string
	Platform string
	OS       string
	Model    string
	Serial   string
	Uptime   string
}

type ConfigurationRetrieved struct {
	IP      string
	Success bool
	Error   string
	SavedTo string
}

type InterfacesRetrieved struct {
	IP    string
	Count int
	Error string
}

type NeighborsRetrieved struct {
	IP    string
	Count int
	Error string
}

type DeviceSaved struct {
	IP           string
	DatabasePath string
	Filename     string
}

type DiscoveryCompleted struct {
	IP       string
	Port     int
	Success  bool
	ErrorMsg string
	Duration time.Duration
}
