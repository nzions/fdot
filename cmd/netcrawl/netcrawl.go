package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nzions/dsjdb"
	"github.com/nzions/fdot/pkg/fdh/credmgr"
	"github.com/nzions/fdot/pkg/fdh/fuser"
	"github.com/nzions/fdot/pkg/fdh/netdevice"
	"github.com/nzions/fdot/pkg/fdh/netssh"
)

// Version is the semantic version of netcrawl
const Version = "1.0.0"

var (
	deviceIP    = flag.String("device", "", "Target device IP address (required)")
	port        = flag.Int("port", 22, "SSH port")
	timeout     = flag.Duration("timeout", 30*time.Second, "Connection timeout")
	showVersion = flag.Bool("version", false, "Show version and exit")
)

// netcrawl connects to network switches via SSH, executes show commands,
// saves output to files, parses the data, and stores it in dsjdb
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command-line flags
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("netcrawl %s\n", Version)
		return nil
	}

	// Validate required flags
	if *deviceIP == "" {
		fmt.Fprintf(os.Stderr, "Error: -device flag is required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: netcrawl -device <ip-address> [options]\n\n")
		flag.PrintDefaults()
		return fmt.Errorf("missing required flag: -device")
	}

	fmt.Println("NetCrawl - Network Device Discovery Tool")
	fmt.Printf("Version: %s\n", Version)
	fmt.Println("=========================================")

	// load ssh creds
	un, pw, err := fuser.CurrentUser.SSHCreds()
	switch err {
	case nil:
		// all good
	case credmgr.ErrNotFound:
		fmt.Println("❌ No SSH credentials found")
		fmt.Println("   Please set them using: credmgr setssh <username> <password>")
		return nil
	default:
		return fmt.Errorf("loading ssh creds: %w", err)
	}

	fmt.Printf("✓ Loaded SSH credentials for user: %s\n", un)
	fmt.Printf("✓ Target device: %s:%d\n", *deviceIP, *port)

	// Create SSH client
	client := netssh.NewClient(netssh.Config{
		Host:     *deviceIP,
		Port:     *port,
		Username: un,
		Password: pw,
		Timeout:  *timeout,
	})

	fmt.Println("→ Connecting to device...")
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", *deviceIP, err)
	}
	defer client.Close()
	fmt.Println("✓ Connected successfully")

	// Create output directory for this device
	deviceDir := filepath.Join(fuser.CurrentUser.NetworkDir, *deviceIP)
	if err := os.MkdirAll(deviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create device directory: %w", err)
	}

	// Step 1: Execute show version to detect device type
	fmt.Println("→ Executing: show version")
	showVersionOutput, err := client.ExecuteCommand("show version")
	if err != nil {
		return fmt.Errorf("failed to execute show version: %w", err)
	}

	// Save show version output
	showVerFile := filepath.Join(deviceDir, "show_version.txt")
	if err := os.WriteFile(showVerFile, []byte(showVersionOutput), 0644); err != nil {
		return fmt.Errorf("failed to save show version output: %w", err)
	}
	fmt.Printf("✓ Saved output to: %s\n", showVerFile)

	// Step 2: Parse show version and create appropriate device instance
	fmt.Println("→ Detecting device type...")
	device, err := netdevice.NewDevice(client, showVersionOutput)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	// Set the IP address
	device.SetIPAddress(*deviceIP)

	fmt.Printf("✓ Detected device:\n")
	fmt.Printf("  Platform:  %s\n", device.GetPlatform())
	fmt.Printf("  OS:        %s\n", device.GetOSVersion())
	fmt.Printf("  Model:     %s\n", device.GetModel())
	fmt.Printf("  Serial:    %s\n", device.GetSerial())
	fmt.Printf("  Uptime:    %s\n", device.GetUptime())

	// Step 3: Get configuration
	fmt.Println("→ Retrieving configuration...")
	config, err := device.GetConfig()
	if err != nil {
		fmt.Printf("⚠ Warning: failed to get config: %v\n", err)
	} else {
		configFile := filepath.Join(deviceDir, "show_running_config.txt")
		if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("✓ Saved config to: %s\n", configFile)
	}

	// Step 4: Get interfaces
	fmt.Println("→ Retrieving interfaces...")
	interfaces, err := device.GetInterfaces()
	if err != nil {
		fmt.Printf("⚠ Warning: failed to get interfaces: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d interfaces\n", len(interfaces))
	}

	// Step 5: Get neighbors
	fmt.Println("→ Retrieving neighbors...")
	neighbors, err := device.GetNeighbors()
	if err != nil {
		fmt.Printf("⚠ Warning: failed to get neighbors: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d neighbors\n", len(neighbors))
	}

	// Step 6: Save device info to database
	fmt.Println("→ Saving to database...")
	deviceInfo := device.GetDeviceInfo()
	deviceInfo.RawOutputDir = deviceDir

	dbPath := filepath.Join(fuser.CurrentUser.DataDir, "devices")
	db, err := dsjdb.NewJSDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Use device IP as the filename
	deviceFile := fmt.Sprintf("%s.json", *deviceIP)
	if err := db.WriteJSONFile(deviceFile, deviceInfo); err != nil {
		return fmt.Errorf("failed to save device to database: %w", err)
	}

	fmt.Printf("✓ Device data saved to database: %s/%s\n", dbPath, deviceFile)
	fmt.Println("\n✅ NetCrawl completed successfully!")

	return nil
}
