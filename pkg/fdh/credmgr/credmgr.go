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
	Version = "2.0.0"
)

// Read retrieves raw credential bytes by name.
func Read(name string) ([]byte, error) {
	return readCredential(name)
}

// Write stores raw credential bytes with the given name.
func Write(name string, data []byte) error {
	return writeCredential(name, data)
}

// ReadKey retrieves a credential key as a string.
func ReadKey(name string) (string, error) {
	data, err := Read(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteKey stores a string credential key.
func WriteKey(name, key string) error {
	return Write(name, []byte(key))
}

// ReadUserCred retrieves a username/password credential.
func ReadUserCred(name string) (*UnPw, error) {
	data, err := Read(name)
	if err != nil {
		return nil, err
	}
	return unmarshalUnPw(data)
}

// WriteUserCred stores a username/password credential.
func WriteUserCred(name string, cred *UnPw) error {
	return Write(name, cred.marshal())
}

// Delete removes a credential by name.
func Delete(name string) error {
	return deleteCredential(name)
}

// List returns all credential names.
func List() ([]string, error) {
	return listCredentials()
}

// Deprecated: Use ReadKey instead.
func ReadString(name string) (string, error) {
	return ReadKey(name)
}

// Deprecated: Use WriteKey instead.
func WriteString(name, value string) error {
	return WriteKey(name, value)
}
