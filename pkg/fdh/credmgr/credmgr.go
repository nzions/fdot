// Package credmgr provides cross-platform credential management functionality.
// It uses Windows Credential Manager on Windows and Linux kernel keyring on Linux.
package credmgr

import "errors"

var (
	// ErrNotFound is returned when a credential is not found
	ErrNotFound = errors.New("credential not found")
	// ErrNotSupported is returned on unsupported platforms
	ErrNotSupported = errors.New("credential manager not supported on this platform")
)

const (
	Version = "1.0.0"
)

// Read retrieves a credential by name
func Read(name string) ([]byte, error) {
	return readCredential(name)
}

// Write stores a credential with the given name and data
func Write(name string, data []byte) error {
	return writeCredential(name, data)
}

// Delete removes a credential by name
func Delete(name string) error {
	return deleteCredential(name)
}

// List returns all available credential names
func List() ([]string, error) {
	return listCredentials()
}

// ReadString is a convenience function to read string credentials
func ReadString(name string) (string, error) {
	data, err := Read(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteString is a convenience function to write string credentials
func WriteString(name, value string) error {
	return Write(name, []byte(value))
}
