// Package main implements a simple credential manager CLI tool.
// Usage:
//
//	credmgr get <name>          - Retrieve credential
//	credmgr set <name> <data>   - Store credential
//	credmgr del <name>          - Delete credential
//	credmgr deletedb            - Delete entire credential database
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
)

const Version = "1.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Create credential manager instance
	cm, err := credmgr.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating credential manager: %v\n", err)
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "get":
		handleGet(cm)
	case "set":
		handleSet(cm)
	case "setssh":
		handleSetSSH(cm)
	case "getssh":
		handleGetSSH(cm)
	case "getbigkey":
		handleGetBigKey(cm)
	case "del", "delete":
		handleDelete(cm)
	case "deletedb", "cleardb", "clear":
		handleDeleteDB(cm)
	case "list", "ls":
		handleList(cm)
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

func handleGet(cm credmgr.CredManager) {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: credential name required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr get <name>\n")
		os.Exit(1)
	}

	name := os.Args[2]

	data, err := cm.ReadKey(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credential '%s': %v\n", name, err)
		os.Exit(1)
	}

	fmt.Print(data) // No newline to make it easier to pipe/use in scripts
}

func handleSet(cm credmgr.CredManager) {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: credential name and data required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr set <name> <data>\n")
		os.Exit(1)
	}

	name := os.Args[2]
	// Join all remaining args as the data (allows spaces in data)
	data := strings.Join(os.Args[3:], " ")

	err := cm.WriteKey(name, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing credential '%s': %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("Credential '%s' stored successfully\n", name)
}

func handleDelete(cm credmgr.CredManager) {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: credential name required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr del <name>\n")
		os.Exit(1)
	}

	name := os.Args[2]

	err := cm.Delete(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting credential '%s': %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("Credential '%s' deleted successfully\n", name)
}

func handleList(cm credmgr.CredManager) {
	names, err := cm.List()
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

func handleDeleteDB(cm credmgr.CredManager) {
	// Prompt for confirmation since this is destructive
	fmt.Print("This will delete ALL credentials from the database. Are you sure? (yes/no): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "yes" && response != "y" {
		fmt.Println("Operation cancelled")
		return
	}

	err := cm.DeleteDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting credential database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Credential database deleted successfully")
}

func handleSetSSH(cm credmgr.CredManager) {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Error: username and password required\n")
		fmt.Fprintf(os.Stderr, "Usage: credmgr setssh <username> <password>\n")
		os.Exit(1)
	}

	username := os.Args[2]
	password := os.Args[3]

	// Store SSH credentials directly using credmgr
	cred := credmgr.NewUnPw(username, password)
	err := cm.WriteUserCred("fdh-user-ssh-creds", cred)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing SSH credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SSH credentials for '%s' stored successfully\n", username)
}

func handleGetBigKey(cm credmgr.CredManager) {
	// Try to read existing big key
	bigKey, err := cm.ReadKey("fdh-user-bigkey")
	if err == nil {
		fmt.Print(bigKey)
		return
	}

	// Create new big key if it doesn't exist
	randomBytes := make([]byte, 128)
	if _, err := rand.Read(randomBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating big key: %v\n", err)
		os.Exit(1)
	}

	bigKey = hex.EncodeToString(randomBytes)
	if err := cm.WriteKey("fdh-user-bigkey", bigKey); err != nil {
		fmt.Fprintf(os.Stderr, "Error storing big key: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(bigKey) // No newline to make it easier to pipe/use in scripts
}

func handleGetSSH(cm credmgr.CredManager) {
	cred, err := cm.ReadUserCred("fdh-user-ssh-creds")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting SSH credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Username: %s\nPassword: %s\n", cred.Username(), cred.Password())
}
