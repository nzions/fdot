//go:build linux

package credmgr

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Linux syscall numbers for keyring operations (x86_64)
const (
	sysAddKey     = 248
	sysRequestKey = 249
	sysKeyctl     = 250
)

// keyctl operations
const (
	keyctlGetKeyringID = 0
	keyctlDescribe     = 6
	keyctlRead         = 11
	keyctlSearch       = 10
	keyctlUnlink       = 9
)

// Special keyring IDs
const (
	keySpecUserKeyring = ^uintptr(3) // -4 as unsigned (0xFFFFFFFC)
)

// readCredential retrieves a credential from Linux kernel keyring
func readCredential(name string) ([]byte, error) {
	keyName := name
	if CredentialPrefix != "" {
		keyName = CredentialPrefix + "-" + name
	}

	// Search for the key in user keyring
	keyNamePtr, err := syscall.BytePtrFromString(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to convert key name: %w", err)
	}

	keyTypePtr, err := syscall.BytePtrFromString("user")
	if err != nil {
		return nil, fmt.Errorf("failed to convert key type: %w", err)
	}

	keyID, _, errno := syscall.Syscall6(sysRequestKey,
		uintptr(unsafe.Pointer(keyTypePtr)),
		uintptr(unsafe.Pointer(keyNamePtr)),
		0, // payload (not used for search)
		keySpecUserKeyring,
		0, 0)

	if errno != 0 {
		return nil, ErrNotFound
	}

	// Read the key data
	buffer := make([]byte, 4096) // Should be enough for most credentials
	bufferPtr := (*byte)(unsafe.Pointer(&buffer[0]))

	size, _, errno := syscall.Syscall6(sysKeyctl,
		keyctlRead,
		keyID,
		uintptr(unsafe.Pointer(bufferPtr)),
		uintptr(len(buffer)),
		0, 0)

	if errno != 0 {
		return nil, fmt.Errorf("failed to read key: %v", errno)
	}

	return buffer[:size], nil
}

// writeCredential stores a credential in Linux kernel keyring
func writeCredential(name string, data []byte) error {
	keyName := name
	if CredentialPrefix != "" {
		keyName = CredentialPrefix + "-" + name
	}

	keyTypePtr, err := syscall.BytePtrFromString("user")
	if err != nil {
		return fmt.Errorf("failed to convert key type: %w", err)
	}

	keyNamePtr, err := syscall.BytePtrFromString(keyName)
	if err != nil {
		return fmt.Errorf("failed to convert key name: %w", err)
	}

	var dataPtr *byte
	if len(data) > 0 {
		dataPtr = &data[0]
	}

	_, _, errno := syscall.Syscall6(sysAddKey,
		uintptr(unsafe.Pointer(keyTypePtr)),
		uintptr(unsafe.Pointer(keyNamePtr)),
		uintptr(unsafe.Pointer(dataPtr)),
		uintptr(len(data)),
		keySpecUserKeyring,
		0)

	if errno != 0 {
		return fmt.Errorf("failed to add key: %v", errno)
	}

	return nil
}

// deleteCredential removes a credential from Linux kernel keyring
func deleteCredential(name string) error {
	keyName := name
	if CredentialPrefix != "" {
		keyName = CredentialPrefix + "-" + name
	}

	// First find the key
	keyNamePtr, err := syscall.BytePtrFromString(keyName)
	if err != nil {
		return fmt.Errorf("failed to convert key name: %w", err)
	}

	keyTypePtr, err := syscall.BytePtrFromString("user")
	if err != nil {
		return fmt.Errorf("failed to convert key type: %w", err)
	}

	keyID, _, errno := syscall.Syscall6(sysRequestKey,
		uintptr(unsafe.Pointer(keyTypePtr)),
		uintptr(unsafe.Pointer(keyNamePtr)),
		0, // payload (not used for search)
		keySpecUserKeyring,
		0, 0)

	if errno != 0 {
		return ErrNotFound
	}

	// Unlink (delete) the key from the keyring
	_, _, errno = syscall.Syscall(sysKeyctl,
		keyctlUnlink,
		keyID,
		keySpecUserKeyring)

	if errno != 0 {
		return fmt.Errorf("failed to delete key: %v", errno)
	}

	return nil
}

// listCredentials retrieves all credential names from Linux kernel keyring
func listCredentials() ([]string, error) {
	// Read /proc/keys to find user keys
	data, err := os.ReadFile("/proc/keys")
	if err != nil {
		return []string{}, nil // Return empty list if we can't read /proc/keys
	}

	return parseCredentialNames(string(data)), nil
}

// Helper function to extract credential names from /proc/keys (if we wanted full implementation)
func parseCredentialNames(procKeysContent string) []string {
	var names []string
	lines := strings.Split(procKeysContent, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 9 {
			// Field 8 contains the key description
			desc := fields[8]
			if CredentialPrefix == "" {
				// No prefix - include all user keys
				if strings.Contains(line, "user") {
					names = append(names, desc)
				}
			} else {
				prefix := CredentialPrefix + "-"
				if strings.HasPrefix(desc, prefix) {
					name := strings.TrimPrefix(desc, prefix)
					names = append(names, name)
				}
			}
		}
	}

	return names
}
