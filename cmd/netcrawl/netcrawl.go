package main

import (
	"fmt"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
	"github.com/nzions/fdot/pkg/fdh/fuser"
)

// ssh to remote switch, show ver, show run, capture output to file
func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
	}
}

func run() error {
	fmt.Println("running...")

	// load ssh creds
	un, pw, err := fuser.CurrentUser.SSHCreds()
	switch err {
	case nil:
		// all good
	case credmgr.ErrNotFound:
		fmt.Println("no ssh credentials found, please set them using 'credmgr setssh <username> <password>'")
		return nil
	default:
		return fmt.Errorf("loading ssh creds: %w", err)
	}

	// just print them for now
	// later we will use them to ssh to a device
	fmt.Println("ssh user:", un)
	fmt.Println("ssh pass:", pw)

	return nil
}
