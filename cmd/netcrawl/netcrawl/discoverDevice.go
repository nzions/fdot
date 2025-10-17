package netcrawl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nzions/dsjdb"
	"github.com/nzions/eventstream"
	"github.com/nzions/fdot/pkg/fdh/credmgr"
	"github.com/nzions/fdot/pkg/fdh/fuser"
	"github.com/nzions/fdot/pkg/fdh/netdevice"
	"github.com/nzions/fdot/pkg/fdh/netssh"
)

func DiscoverDevice(ctx context.Context, deviceIP *string, port *int, timeout *time.Duration) error {
	log := eventstream.GetFromContext(ctx)

	// load ssh creds
	cred, err := fuser.CurrentUser.SSHCreds()
	switch err {
	case nil:
		// all good
	case credmgr.ErrNotFound:
		log.Errorf("No SSH credentials found - please set them using: credmgr setssh <username> <password>")
		log.Send(DiscoveryCompleted{
			IP:       *deviceIP,
			Port:     *port,
			Success:  false,
			ErrorMsg: "No SSH credentials found",
		})
		return nil
	default:
		return fmt.Errorf("loading ssh creds: %w", err)
	}

	log.Send(DiscoveryStarted{
		IP:       *deviceIP,
		Port:     *port,
		Username: cred.Username(),
	})

	// Create SSH client, and exec show ver
	client := netssh.NewClient(ctx, netssh.Config{
		Host:        *deviceIP,
		Port:        *port,
		Credentials: cred,
		Timeout:     *timeout,
	})
	showVersionOutput, err := client.ExecuteCommand("show version")
	if err != nil {
		return fmt.Errorf("executing show version: %w", err)
	}

	// Create output directory for this device
	deviceDir := filepath.Join(fuser.CurrentUser.NetworkDir, *deviceIP)
	if err := os.MkdirAll(deviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create device directory: %w", err)
	}

	// Save show version output
	showVerFile := filepath.Join(deviceDir, "show_version.txt")
	if err := os.WriteFile(showVerFile, []byte(showVersionOutput), 0644); err != nil {
		return fmt.Errorf("failed to save show version output: %w", err)
	}

	log.Send(ShowVersionRetrieved{
		IP:           *deviceIP,
		OutputLength: len(showVersionOutput),
		SavedTo:      showVerFile,
	})

	// Step 2: Parse show version and create appropriate device instance
	log.Infof("Detecting device type...")
	device, err := netdevice.NewDevice(client, showVersionOutput)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	// Set the IP address
	device.SetIPAddress(*deviceIP)

	log.Send(DeviceDetected{
		IP:       *deviceIP,
		Platform: device.GetPlatform(),
		OS:       device.GetOSVersion(),
		Model:    device.GetModel(),
		Serial:   device.GetSerial(),
		Uptime:   device.GetUptime(),
	})

	// Step 3: Get configuration
	log.Infof("Retrieving configuration...")
	config, err := device.GetConfig()
	if err != nil {
		log.Warnf("Failed to get config: %v", err)
		log.Send(ConfigurationRetrieved{
			IP:      *deviceIP,
			Success: false,
			Error:   err.Error(),
		})
	} else {
		configFile := filepath.Join(deviceDir, "show_running_config.txt")
		if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		log.Send(ConfigurationRetrieved{
			IP:      *deviceIP,
			Success: true,
			SavedTo: configFile,
		})
	}

	// Step 4: Get interfaces
	log.Infof("Retrieving interfaces...")
	interfaces, err := device.GetInterfaces()
	if err != nil {
		log.Warnf("Failed to get interfaces: %v", err)
		log.Send(InterfacesRetrieved{
			IP:    *deviceIP,
			Count: 0,
			Error: err.Error(),
		})
	} else {
		log.Send(InterfacesRetrieved{
			IP:    *deviceIP,
			Count: len(interfaces),
		})
	}

	// Step 5: Get neighbors
	log.Infof("Retrieving neighbors...")
	neighbors, err := device.GetNeighbors()
	if err != nil {
		log.Warnf("Failed to get neighbors: %v", err)
		log.Send(NeighborsRetrieved{
			IP:    *deviceIP,
			Count: 0,
			Error: err.Error(),
		})
	} else {
		log.Send(NeighborsRetrieved{
			IP:    *deviceIP,
			Count: len(neighbors),
		})
	}

	// Step 6: Save device info to database
	log.Infof("Saving to database...")
	deviceInfo := device.GetDeviceInfo()
	deviceInfo.RawOutputDir = deviceDir

	dbPath := filepath.Join(fuser.CurrentUser.DataDir, "devices")
	db, err := dsjdb.NewJSDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Use device IP as the filename
	deviceFile := fmt.Sprintf("%s.json", *deviceIP)
	if err := db.Write(deviceFile, deviceInfo); err != nil {
		return fmt.Errorf("failed to save device to database: %w", err)
	}

	log.Send(DeviceSaved{
		IP:           *deviceIP,
		DatabasePath: dbPath,
		Filename:     deviceFile,
	})

	log.Send(DiscoveryCompleted{
		IP:      *deviceIP,
		Port:    *port,
		Success: true,
	})

	return nil
}
