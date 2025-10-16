//go:build linux

package credmgr

import (
	"fmt"
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
	keyctlRead        = 11
	keyctlUnlink      = 9
	keyctlDescribe    = 6
	keyctlSearch      = 10
	keyctlReadAllKeys = 3
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
	// Use KEYCTL_READ to get all key IDs from the user keyring
	buffer := make([]byte, 4096)
	bufferPtr := (*byte)(unsafe.Pointer(&buffer[0]))

	size, _, errno := syscall.Syscall6(sysKeyctl,
		keyctlReadAllKeys,
		keySpecUserKeyring,
		uintptr(unsafe.Pointer(bufferPtr)),
		uintptr(len(buffer)),
		0, 0)

	if errno != 0 {
		return []string{}, nil // Return empty list if we can't read keyring
	}

	// Parse the returned key IDs (they are returned as int32 array)
	keyCount := int(size) / 4 // Each key ID is 4 bytes
	var names []string

	for i := 0; i < keyCount; i++ {
		keyID := *(*int32)(unsafe.Pointer(&buffer[i*4]))

		// Get the key description using KEYCTL_DESCRIBE
		descBuffer := make([]byte, 256)
		descPtr := (*byte)(unsafe.Pointer(&descBuffer[0]))

		descSize, _, errno := syscall.Syscall6(sysKeyctl,
			keyctlDescribe,
			uintptr(keyID),
			uintptr(unsafe.Pointer(descPtr)),
			uintptr(len(descBuffer)),
			0, 0)

		if errno != 0 {
			continue // Skip keys we can't describe
		}

		// Parse the description: "type;uid;gid;perm;description"
		desc := string(descBuffer[:descSize-1]) // Remove null terminator
		parts := strings.Split(desc, ";")
		if len(parts) >= 5 && parts[0] == "user" {
			keyName := parts[4]

			if CredentialPrefix == "" {
				// No prefix - include all user keys
				names = append(names, keyName)
			} else {
				prefix := CredentialPrefix + "-"
				if after, ok := strings.CutPrefix(keyName, prefix); ok {
					name := after
					names = append(names, name)
				}
			}
		}
	}

	return names, nil
}
