//go:build linux

// Package credmgr provides secure credential storage using Linux kernel keyrings.
//
// # Linux Kernel Keyring Permission Model
//
// The Linux kernel keyring uses a unique permission model with 4 access levels:
//   - possessor: Has the key reachable from their session keyring
//   - owner:     Matches the key's owner UID
//   - group:     Matches the key's group GID
//   - other:     Everyone else
//
// Default key permissions are typically: alswrv-----v------------
//   - possessor: alswrv (all permissions: alter, link, search, write, read, view)
//   - owner:     -----v (view only - can see it exists but cannot read content)
//   - group:     ------ (no access)
//   - other:     ------ (no access)
//
// # The "Possession" Concept
//
// IMPORTANT: Matching the owner UID is NOT sufficient to read a key!
// You need "possession", which is granted when a key is reachable from your
// process's session keyring (@s). This is THE fundamental concept of kernel keyrings.
//
// When you run `keyctl show` (without arguments, showing @s), you'll see your
// session keyring contains a link to your user keyring, shown as "_uid.XXX":
//
//	Session Keyring
//	1046546095 --alswrv   1000   100  keyring: _ses
//	 348545926 --alswrv   1000 65534   \_ keyring: _uid.1000
//	 987654321 --alswrv   1000  1000       \_ user: my-credential
//
// That _uid.1000 link is what grants you "possession" of keys stored in @u.
// The kernel searches from @s → @u, and if the link exists, you "possess" the keys.
//
// # Why This Package Links Keyrings
//
// This package stores credentials in the user keyring (@u) for persistence, but
// explicitly links @u into the session keyring (@s) to ensure "possession".
// Without this link, even though you own the key (same UID), you cannot read it
// because you lack "possession" permission.
//
// This is standard Linux kernel keyring behavior, documented in:
//   - keyctl(2) man page: https://man7.org/linux/man-pages/man2/keyctl.2.html
//   - keyrings(7) man page: https://man7.org/linux/man-pages/man7/keyrings.7.html
//   - Stack Overflow explanation: https://stackoverflow.com/a/79389296
//
// Reference implementation: PAM modules (like pam_keyinit) automatically create
// this link when you log in, which is why `keyctl add user foo bar @u` followed
// by `keyctl read <keyid>` typically works in normal shells.
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
	keyctlLink        = 8
)

// Special keyring IDs
const (
	keySpecSessionKeyring = ^uintptr(2) // -3 as unsigned (0xFFFFFFFD)
	keySpecUserKeyring    = ^uintptr(3) // -4 as unsigned (0xFFFFFFFC)
)

// ensureUserKeyringLinked links the user keyring (@u) to the session keyring (@s).
// This grants "possession" permission, allowing us to read/write/delete keys.
// See package documentation and https://stackoverflow.com/a/79389296 for details.
func ensureUserKeyringLinked() {
	syscall.Syscall(sysKeyctl, keyctlLink, keySpecUserKeyring, keySpecSessionKeyring)
	// Errors ignored - already linked or insufficient permissions
}

// requestKey finds a key by name and returns its ID.
func requestKey(name string) (uintptr, error) {
	keyTypePtr, err := syscall.BytePtrFromString("user")
	if err != nil {
		return 0, fmt.Errorf("failed to convert key type: %w", err)
	}

	keyNamePtr, err := syscall.BytePtrFromString(name)
	if err != nil {
		return 0, fmt.Errorf("failed to convert key name: %w", err)
	}

	ensureUserKeyringLinked()

	keyID, _, errno := syscall.Syscall6(sysRequestKey,
		uintptr(unsafe.Pointer(keyTypePtr)),
		uintptr(unsafe.Pointer(keyNamePtr)),
		0, // no callout info
		0, // search default path: @s → @u
		0, 0)

	if errno != 0 {
		return 0, ErrNotFound
	}

	return keyID, nil
}

// readCredential retrieves a credential from Linux kernel keyring.
func readCredential(name string) ([]byte, error) {
	keyID, err := requestKey(name)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 4096)
	size, _, errno := syscall.Syscall6(sysKeyctl,
		keyctlRead,
		keyID,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		0, 0)

	if errno != 0 {
		return nil, fmt.Errorf("failed to read key: %v", errno)
	}

	return buffer[:size], nil
}

// writeCredential stores a credential in Linux kernel keyring.
func writeCredential(name string, data []byte) error {
	keyTypePtr, err := syscall.BytePtrFromString("user")
	if err != nil {
		return fmt.Errorf("failed to convert key type: %w", err)
	}

	keyNamePtr, err := syscall.BytePtrFromString(name)
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

// deleteCredential removes a credential from Linux kernel keyring.
func deleteCredential(name string) error {
	keyID, err := requestKey(name)
	if err != nil {
		return err
	}

	_, _, errno := syscall.Syscall(sysKeyctl,
		keyctlUnlink,
		keyID,
		keySpecUserKeyring)

	if errno != 0 {
		return fmt.Errorf("failed to delete key: %v", errno)
	}

	return nil
}

// listCredentials retrieves all credential names from Linux kernel keyring.
func listCredentials() ([]string, error) {
	buffer := make([]byte, 4096)
	size, _, errno := syscall.Syscall6(sysKeyctl,
		keyctlRead,
		keySpecUserKeyring,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		0, 0)

	if errno != 0 {
		return []string{}, nil
	}

	keyCount := int(size) / 4 // Each key ID is 4 bytes (int32)
	names := make([]string, 0, keyCount)

	for i := 0; i < keyCount; i++ {
		keyID := *(*int32)(unsafe.Pointer(&buffer[i*4]))

		descBuffer := make([]byte, 256)
		descSize, _, errno := syscall.Syscall6(sysKeyctl,
			keyctlDescribe,
			uintptr(keyID),
			uintptr(unsafe.Pointer(&descBuffer[0])),
			uintptr(len(descBuffer)),
			0, 0)

		if errno != 0 {
			continue
		}

		// Parse description format: "type;uid;gid;perm;description"
		desc := string(descBuffer[:descSize-1])
		parts := strings.Split(desc, ";")
		if len(parts) >= 5 && parts[0] == "user" {
			names = append(names, parts[4])
		}
	}

	return names, nil
}
