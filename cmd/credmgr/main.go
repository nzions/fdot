// Package main implements a simple credential manager CLI tool.
// Usage:
//
//	credmgr get <name>          - Retrieve credential
//	credmgr set <name> <data>   - Store credential
//	credmgr del <name>          - Delete credential
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "get":
		handleGet()
	case "set":
		handleSet()
	case "del", "delete":
		handleDelete()
	case "list", "ls":
		handleList()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("credmgr - Simple Windows Credential Manager CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  credmgr get <name>          Retrieve credential")
	fmt.Println("  credmgr set <name> <data>   Store credential")
	fmt.Println("  credmgr del <name>          Delete credential")
	fmt.Println("  credmgr list                List all credentials")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  credmgr set myapp-token secret123")
	fmt.Println("  credmgr get myapp-token")
	fmt.Println("  credmgr del myapp-token")
}

func handleGet() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: credential name required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr get <name>\n")
		os.Exit(1)
	}

	name := os.Args[2]

	data, err := credmgr.ReadString(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credential '%s': %v\n", name, err)
		os.Exit(1)
	}

	fmt.Print(data) // No newline to make it easier to pipe/use in scripts
}

func handleSet() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: credential name and data required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr set <name> <data>\n")
		os.Exit(1)
	}

	name := os.Args[2]
	// Join all remaining args as the data (allows spaces in data)
	data := strings.Join(os.Args[3:], " ")

	err := credmgr.WriteString(name, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing credential '%s': %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("Credential '%s' stored successfully\n", name)
}

func handleDelete() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: credential name required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr del <name>\n")
		os.Exit(1)
	}

	name := os.Args[2]

	err := credmgr.Delete(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting credential '%s': %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("Credential '%s' deleted successfully\n", name)
}

func handleList() {
	names, err := credmgr.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing credentials: %v\n", err)
		os.Exit(1)
	}

	if len(names) == 0 {
		fmt.Println("No credentials found")
		return
	}

	for _, name := range names {
		fmt.Println(name)
	}
}
