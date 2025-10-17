// Example demonstrating the new credmgr v2.0.0 API
package main

import (
	"fmt"
	"log"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
)

func main() {
	fmt.Println("=== credmgr API v2.0.0 Demo ===\n")

	// 1. Raw bytes API - Read/Write
	fmt.Println("1. Raw bytes API (Read/Write):")
	rawData := []byte("binary-data-12345")
	if err := credmgr.Write("raw-credential", rawData); err != nil {
		log.Fatalf("Write failed: %v", err)
	}
	retrieved, err := credmgr.Read("raw-credential")
	if err != nil {
		log.Fatalf("Read failed: %v", err)
	}
	fmt.Printf("   Stored: %q\n", rawData)
	fmt.Printf("   Retrieved: %q\n", retrieved)
	fmt.Println("   ✓ Raw bytes work")

	// 2. Key/token API - ReadKey/WriteKey
	fmt.Println("\n2. Key/token API (ReadKey/WriteKey):")
	apiKey := "sk-proj-1234567890abcdef"
	if err := credmgr.WriteKey("api-key-demo", apiKey); err != nil {
		log.Fatalf("WriteKey failed: %v", err)
	}
	retrievedKey, err := credmgr.ReadKey("api-key-demo")
	if err != nil {
		log.Fatalf("ReadKey failed: %v", err)
	}
	fmt.Printf("   Stored: %s\n", apiKey)
	fmt.Printf("   Retrieved: %s\n", retrievedKey)
	fmt.Println("   ✓ Key storage works")

	// 3. Username/Password API - ReadUserCred/WriteUserCred
	fmt.Println("\n3. Username/Password API (ReadUserCred/WriteUserCred):")
	cred := credmgr.NewUnPw("john.doe", "secretpass123")
	if err := credmgr.WriteUserCred("user-cred-demo", cred); err != nil {
		log.Fatalf("WriteUserCred failed: %v", err)
	}
	retrievedCred, err := credmgr.ReadUserCred("user-cred-demo")
	if err != nil {
		log.Fatalf("ReadUserCred failed: %v", err)
	}
	fmt.Printf("   Stored: username=%s, password=%s\n", cred.Username(), cred.Password())
	fmt.Printf("   Retrieved: username=%s, password=%s\n", retrievedCred.Username(), retrievedCred.Password())
	fmt.Println("   ✓ UnPw credentials work")

	// 4. List all credentials
	fmt.Println("\n4. List all credentials:")
	names, err := credmgr.List()
	if err != nil {
		log.Fatalf("List failed: %v", err)
	}
	for _, name := range names {
		fmt.Printf("   - %s\n", name)
	}

	// Cleanup
	fmt.Println("\n5. Cleanup:")
	for _, name := range []string{"raw-credential", "api-key-demo", "user-cred-demo"} {
		if err := credmgr.Delete(name); err != nil {
			log.Printf("Warning: failed to delete %s: %v", name, err)
		} else {
			fmt.Printf("   ✓ Deleted: %s\n", name)
		}
	}

	fmt.Println("\n=== Demo Complete ===")
}
