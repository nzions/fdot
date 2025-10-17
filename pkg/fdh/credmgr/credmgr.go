// Package credmgr provides cross-platform credential management.
// Uses Windows Credential Manager on Windows and AES-encrypted file storage on Linux.
package credmgr

import (
	"errors"
)

var (
	// ErrNotFound is returned when a credential is not found.
	ErrNotFound = errors.New("credential not found")
	// ErrNotSupported is returned on unsupported platforms.
	ErrNotSupported = errors.New("credential manager not supported on this platform")
	// ErrInvalidFormat is returned when a credential has invalid format.
	ErrInvalidFormat = errors.New("invalid credential format")
)

const (
	// Version is the credmgr package version.
	Version = "3.0.0"
)

// CredManager defines the interface for credential management operations.
type CredManager interface {
	// Read retrieves raw credential bytes by name.
	Read(name string) ([]byte, error)

	// Write stores raw credential bytes with the given name.
	Write(name string, data []byte) error

	// ReadKey retrieves a credential key as a string.
	ReadKey(name string) (string, error)

	// WriteKey stores a string credential key.
	WriteKey(name, key string) error

	// ReadUserCred retrieves a username/password credential.
	ReadUserCred(name string) (UserCred, error)

	// WriteUserCred stores a username/password credential.
	WriteUserCred(name string, cred UserCred) error

	// Delete removes a credential by name.
	Delete(name string) error

	// DeleteDB removes the entire credential database.
	DeleteDB() error

	// List returns all credential names.
	List() ([]string, error)
}

// New creates a new CredManager with the specified storage path.
//
// Path behavior:
//   - Empty string ("") or nil: Uses platform default storage
//   - Windows: Uses Windows Credential Manager
//   - Linux: Uses default file path (~/.fdot/credentials.enc)
//   - Non-empty string: Uses disk-based storage at specified path
//   - All platforms: AES-encrypted file storage at the given path
//
// Examples:
//
//	credmgr := credmgr.New("")                    // Platform default
//	credmgr := credmgr.New("/custom/creds.enc")   // Custom file path
func New(path string) (CredManager, error) {
	return newCredManager(path)
}

// Default returns a CredManager using the platform's default storage mechanism.
//   - Windows: Windows Credential Manager
//   - Linux: ~/.local/credmgr/credentials.enc
//   - Other: Returns error for unsupported operations
func Default() (CredManager, error) {
	return defaultCredManager()
}
