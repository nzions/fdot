package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/nzions/eventstream"
	"github.com/nzions/fdot/cmd/netcrawl/netcrawl"
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

	log := eventstream.DefaultHandler
	ctx := eventstream.AddToContext(context.Background(), log)
	if err := netcrawl.DiscoverDevice(ctx, deviceIP, port, timeout); err != nil {
		return fmt.Errorf("discovering device: %w", err)
	}
	return nil
}
