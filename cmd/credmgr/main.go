// Package main implements a simple credential manager CLI tool.
// Usage:
//
//	credmgr get <name>          - Retrieve credential
//	credmgr set <name> <data>   - Store credential
//	credmgr del <name>          - Delete credential
//	credmgr deletedb            - Delete entire credential database
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
	"github.com/nzions/fdot/pkg/fdh/fuser"
)

const Version = "1.1.0"

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
	case "setssh":
		handleSetSSH()
	case "getssh":
		handleGetSSH()
	case "getbigkey":
		handleGetBigKey()
	case "del", "delete":
		handleDelete()
	case "deletedb", "cleardb", "clear":
		handleDeleteDB()
	case "list", "ls":
		handleList()
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("credmgr - Simple Credential Manager CLI")
	fmt.Printf("Binary Version   %s\n", Version)
	fmt.Printf("Library Version  %s\n", credmgr.Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  credmgr get <name>          Retrieve credential")
	fmt.Println("  credmgr set <name> <data>   Store credential")
	fmt.Println("  credmgr setssh <un> <pw>    Store SSH credentials")
	fmt.Println("  credmgr getssh              Get SSH credentials")
	fmt.Println("  credmgr getbigkey           Get or create big key")
	fmt.Println("  credmgr del <name>          Delete credential")
	fmt.Println("  credmgr deletedb            Delete ALL credentials (with confirmation)")
	fmt.Println("  credmgr list                List all credentials")
	fmt.Println("  credmgr version             Show version information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  credmgr set myapp-token secret123")
	fmt.Println("  credmgr setssh john mypassword")
	fmt.Println("  credmgr getssh")
	fmt.Println("  credmgr getbigkey")
	fmt.Println("  credmgr get myapp-token")
	fmt.Println("  credmgr del myapp-token")
}

func printVersion() {
	fmt.Println(Version)
}

func handleGet() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: credential name required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr get <name>\n")
		os.Exit(1)
	}

	name := os.Args[2]

	data, err := credmgr.ReadKey(name)
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

	err := credmgr.WriteKey(name, data)
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

func handleDeleteDB() {
	// Prompt for confirmation since this is destructive
	fmt.Print("This will delete ALL credentials from the database. Are you sure? (yes/no): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "yes" && response != "y" {
		fmt.Println("Operation cancelled")
		return
	}

	err := credmgr.DeleteDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting credential database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Credential database deleted successfully")
}

func handleSetSSH() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: username and password required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr setssh <username> <password>\n")
		os.Exit(1)
	}

	username := os.Args[2]
	password := os.Args[3]

	err := fuser.CurrentUser.SetSSHCreds(username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing SSH credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SSH credentials for '%s' stored successfully\n", username)
}

func handleGetBigKey() {
	bigKey, err := fuser.CurrentUser.BigKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting big key: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(bigKey) // No newline to make it easier to pipe/use in scripts
}

func handleGetSSH() {
	cred, err := fuser.CurrentUser.SSHCreds()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting SSH credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Username: %s\nPassword: %s\n", cred.Username(), cred.Password())
}
