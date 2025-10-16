//go:build windows

package credmgr

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	advapi32           = syscall.NewLazyDLL("advapi32.dll")
	procCredReadW      = advapi32.NewProc("CredReadW")
	procCredWriteW     = advapi32.NewProc("CredWriteW")
	procCredDeleteW    = advapi32.NewProc("CredDeleteW")
	procCredEnumerateW = advapi32.NewProc("CredEnumerateW")
	procCredFree       = advapi32.NewProc("CredFree")
)

const (
	// CRED_TYPE_GENERIC represents a generic credential
	credTypeGeneric = 1
	// CRED_PERSIST_LOCAL_MACHINE persists for all sessions on local machine
	credPersistLocalMachine = 2
)

// credential represents the Windows CREDENTIAL structure
type credential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

// utf16PtrToString converts a UTF16 pointer to a Go string
func utf16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}

	// Find length by looking for null terminator
	length := 0
	for {
		if *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(length*2))) == 0 {
			break
		}
		length++
	}

	if length == 0 {
		return ""
	}

	// Convert to slice and then to string
	utf16Slice := (*[1 << 20]uint16)(unsafe.Pointer(ptr))[:length:length]
	return syscall.UTF16ToString(utf16Slice)
}

// readCredential retrieves a credential from Windows Credential Manager
func readCredential(name string) ([]byte, error) {
	targetName := name
	if CredentialPrefix != "" {
		targetName = CredentialPrefix + "-" + name
	}
	targetNamePtr, err := syscall.UTF16PtrFromString(targetName)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target name: %w", err)
	}

	var cred *credential
	ret, _, _ := procCredReadW.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		uintptr(credTypeGeneric),
		0, // flags (reserved)
		uintptr(unsafe.Pointer(&cred)),
	)

	if ret == 0 {
		return nil, ErrNotFound
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(cred)))

	if cred.CredentialBlobSize == 0 {
		return []byte{}, nil
	}

	// Copy the credential data
	data := make([]byte, cred.CredentialBlobSize)
	copy(data, (*[1 << 30]byte)(unsafe.Pointer(cred.CredentialBlob))[:cred.CredentialBlobSize:cred.CredentialBlobSize])

	return data, nil
}

// writeCredential stores a credential in Windows Credential Manager
func writeCredential(name string, data []byte) error {
	targetName := name
	if CredentialPrefix != "" {
		targetName = CredentialPrefix + "-" + name
	}
	targetNamePtr, err := syscall.UTF16PtrFromString(targetName)
	if err != nil {
		return fmt.Errorf("failed to convert target name: %w", err)
	}

	var dataPtr *byte
	if len(data) > 0 {
		dataPtr = &data[0]
	}

	cred := credential{
		Type:               credTypeGeneric,
		TargetName:         targetNamePtr,
		CredentialBlobSize: uint32(len(data)),
		CredentialBlob:     dataPtr,
		Persist:            credPersistLocalMachine,
	}

	ret, _, _ := procCredWriteW.Call(
		uintptr(unsafe.Pointer(&cred)),
		0, // flags (reserved)
	)

	if ret == 0 {
		return fmt.Errorf("failed to write credential: %s", targetName)
	}

	return nil
}

// deleteCredential removes a credential from Windows Credential Manager
func deleteCredential(name string) error {
	targetName := name
	if CredentialPrefix != "" {
		targetName = CredentialPrefix + "-" + name
	}
	targetNamePtr, err := syscall.UTF16PtrFromString(targetName)
	if err != nil {
		return fmt.Errorf("failed to convert target name: %w", err)
	}

	ret, _, _ := procCredDeleteW.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		uintptr(credTypeGeneric),
		0, // flags (reserved)
	)

	if ret == 0 {
		return ErrNotFound
	}

	return nil
}

// listCredentials retrieves all generic credentials from Windows Credential Manager
func listCredentials() ([]string, error) {
	var count uint32
	var creds **credential

	ret, _, _ := procCredEnumerateW.Call(
		0, // filter (null for all)
		0, // flags (reserved)
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&creds)),
	)

	if ret == 0 {
		return []string{}, nil // No credentials found
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(creds)))

	// Convert array of credential pointers to slice
	credSlice := (*[1 << 20]*credential)(unsafe.Pointer(creds))[:count:count]

	var names []string
	for _, cred := range credSlice {
		if cred.Type == credTypeGeneric {
			fullName := utf16PtrToString(cred.TargetName)
			if CredentialPrefix == "" {
				names = append(names, fullName)
			} else {
				prefix := CredentialPrefix + "-"
				if strings.HasPrefix(fullName, prefix) {
					name := strings.TrimPrefix(fullName, prefix)
					names = append(names, name)
				}
			}
		}
	}

	return names, nil
}
