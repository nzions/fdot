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
	credTypeGeneric         = 1
	credPersistLocalMachine = 2
)

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

// windowsCredManager implements CredManager for Windows using Windows Credential Manager
type windowsCredManager struct {
	// Windows Credential Manager doesn't need a file path
	// All credentials are stored in the system's credential store
}

// diskCredManager implements CredManager for any platform using AES-encrypted file storage
type diskCredManager struct {
	credFilePath string
	// This will be the same as linuxCredManager but available on Windows too
	// Implementation will be similar to Linux version but without platform restrictions
}

// newCredManager creates a new CredManager for Windows
func newCredManager(path string) (CredManager, error) {
	if path == "" {
		// Use Windows Credential Manager (default)
		return &windowsCredManager{}, nil
	}

	// Use disk-based storage at specified path
	return &diskCredManager{
		credFilePath: path,
	}, nil
}

// defaultCredManager returns the default CredManager for Windows (Windows Credential Manager)
func defaultCredManager() (CredManager, error) {
	return &windowsCredManager{}, nil
}

// utf16PtrToString converts a UTF16 pointer to a Go string
func utf16PtrToString(ptr *uint16) string {
	if ptr == nil {
		return ""
	}

	// Find the length
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
	slice := (*[1 << 20]uint16)(unsafe.Pointer(ptr))[:length:length]
	return syscall.UTF16ToString(slice)
}

// Windows Credential Manager implementation

// Read retrieves raw credential bytes by name.
func (wm *windowsCredManager) Read(name string) ([]byte, error) {
	targetNamePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target name: %w", err)
	}

	var credPtr *credential
	ret, _, _ := procCredReadW.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		uintptr(credTypeGeneric),
		0,
		uintptr(unsafe.Pointer(&credPtr)),
	)

	if ret == 0 {
		return nil, ErrNotFound
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(credPtr)))

	if credPtr.CredentialBlobSize == 0 {
		return []byte{}, nil
	}

	data := (*[1 << 20]byte)(unsafe.Pointer(credPtr.CredentialBlob))[:credPtr.CredentialBlobSize:credPtr.CredentialBlobSize]
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// Write stores raw credential bytes with the given name.
func (wm *windowsCredManager) Write(name string, data []byte) error {
	targetNamePtr, err := syscall.UTF16PtrFromString(name)
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
		0,
	)

	if ret == 0 {
		return fmt.Errorf("failed to write credential: %s", name)
	}

	return nil
}

// ReadKey retrieves a credential key as a string.
func (wm *windowsCredManager) ReadKey(name string) (string, error) {
	data, err := wm.Read(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteKey stores a string credential key.
func (wm *windowsCredManager) WriteKey(name, key string) error {
	return wm.Write(name, []byte(key))
}

// ReadUserCred retrieves a username/password credential.
func (wm *windowsCredManager) ReadUserCred(name string) (UserCred, error) {
	data, err := wm.Read(name)
	if err != nil {
		return nil, err
	}
	return unmarshalUnPw(data)
}

// WriteUserCred stores a username/password credential.
func (wm *windowsCredManager) WriteUserCred(name string, cred UserCred) error {
	// Type assert to access marshal method
	if uc, ok := cred.(*obfuscatedUserCred); ok {
		return wm.Write(name, uc.marshal())
	}
	// Fallback: reconstruct from interface
	reconstructed := newObfuscatedUserCred(cred.Username(), cred.Password())
	return wm.Write(name, reconstructed.marshal())
}

// Delete removes a credential by name.
func (wm *windowsCredManager) Delete(name string) error {
	targetNamePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("failed to convert target name: %w", err)
	}

	ret, _, _ := procCredDeleteW.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		uintptr(credTypeGeneric),
		0,
	)

	if ret == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteDB removes all generic credentials from Windows Credential Manager.
func (wm *windowsCredManager) DeleteDB() error {
	// First get all credential names
	names, err := wm.List()
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	// Delete each credential individually
	var errors []error
	for _, name := range names {
		if err := wm.Delete(name); err != nil {
			// Continue deleting others even if one fails
			errors = append(errors, fmt.Errorf("failed to delete credential %q: %w", name, err))
		}
	}

	// Return combined errors if any occurred
	if len(errors) > 0 {
		var errStr strings.Builder
		errStr.WriteString("failed to delete some credentials:")
		for _, e := range errors {
			errStr.WriteString("\n  ")
			errStr.WriteString(e.Error())
		}
		return fmt.Errorf(errStr.String())
	}

	return nil
}

// List returns all credential names.
func (wm *windowsCredManager) List() ([]string, error) {
	var count uint32
	var creds **credential

	ret, _, _ := procCredEnumerateW.Call(
		0,
		0,
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&creds)),
	)

	if ret == 0 {
		return []string{}, nil
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(creds)))

	credSlice := (*[1 << 20]*credential)(unsafe.Pointer(creds))[:count:count]
	names := make([]string, 0, count)

	for _, cred := range credSlice {
		if cred.Type == credTypeGeneric {
			names = append(names, utf16PtrToString(cred.TargetName))
		}
	}

	return names, nil
}

// TODO: Implement diskCredManager methods for when a specific path is provided on Windows
// For now, we'll implement basic stubs that return ErrNotSupported

// Disk-based storage implementation (when path is specified on Windows)

func (dm *diskCredManager) Read(name string) ([]byte, error) {
	return nil, ErrNotSupported // TODO: Implement AES file storage
}

func (dm *diskCredManager) Write(name string, data []byte) error {
	return ErrNotSupported // TODO: Implement AES file storage
}

func (dm *diskCredManager) ReadKey(name string) (string, error) {
	return "", ErrNotSupported
}

func (dm *diskCredManager) WriteKey(name, key string) error {
	return ErrNotSupported
}

func (dm *diskCredManager) ReadUserCred(name string) (UserCred, error) {
	return nil, ErrNotSupported
}

func (dm *diskCredManager) WriteUserCred(name string, cred UserCred) error {
	return ErrNotSupported
}

func (dm *diskCredManager) Delete(name string) error {
	return ErrNotSupported
}

func (dm *diskCredManager) DeleteDB() error {
	return ErrNotSupported
}

func (dm *diskCredManager) List() ([]string, error) {
	return nil, ErrNotSupported
}
