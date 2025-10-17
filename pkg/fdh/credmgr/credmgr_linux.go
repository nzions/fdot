//go:build linux

// Package credmgr provides credential management on Linux using AES-encrypted file storage.
//
// # Storage Architecture
//
// Credentials are stored in an AES-256-GCM encrypted file:
//   - Location: ~/.local/share/fdot/credentials.enc
//   - Format: JSON map encrypted with AES-256-GCM
//   - Permissions: 0600 (owner read/write only)
//
// # Encryption Key Source
//
// The encryption key MUST be provided via the CREDMGR_KEY environment variable:
//   - Format: 64 hex characters (32 bytes)
//   - Example: export CREDMGR_KEY="0123456789abcdef..."
//   - Generate: openssl rand -hex 32
//
// If CREDMGR_KEY is not set or invalid, credential operations will fail.
//
// # Security Model
//
// This provides file-based credential persistence with these properties:
//   - Encrypted at rest (AES-256-GCM with authentication)
//   - Per-user isolation (file permissions 0600)
//   - Key management is user's responsibility
//   - Protection level: Similar to Windows Credential Manager
//
// Does NOT protect against:
//   - Root/administrator access
//   - Attackers who obtain both the encrypted file AND the key
//   - Memory dumps while credentials are in use
//
// # Design Rationale
//
// Previous implementation used Linux kernel keyrings, which provide excellent
// security but do NOT persist across reboots (they are RAM-only). This file-based
// approach trades some security for persistence, matching Windows behavior.
//
// For development/convenience credentials (API keys, tokens, etc.), this provides
// a reasonable balance. For high-security credentials, consider using kernel keyrings
// (session-only) or prompting for a master password.
//
// See docs/linux-kernel-keyring.bak/ for the archived kernel keyring implementation.
package credmgr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/nzions/fdot/pkg/fdotconfig"
)

var (
	// In-memory cache of decrypted credentials
	credCache      map[string][]byte
	credCacheMutex sync.RWMutex
	credCacheInit  sync.Once

	// Cached encryption key
	encryptionKey []byte
	keyInitOnce   sync.Once
	keyInitError  error
)

// getEncryptionKey loads and validates the encryption key from environment variable
func getEncryptionKey() ([]byte, error) {
	keyInitOnce.Do(func() {
		keyHex := os.Getenv(fdotconfig.CredMgrEnvVar)
		if keyHex == "" {
			keyInitError = fmt.Errorf("%s environment variable not set", fdotconfig.CredMgrEnvVar)
			return
		}

		key, err := hex.DecodeString(keyHex)
		if err != nil {
			keyInitError = fmt.Errorf("invalid %s format (expected 64 hex chars): %w", fdotconfig.CredMgrEnvVar, err)
			return
		}

		if len(key) != 32 {
			keyInitError = fmt.Errorf("invalid %s length (expected 32 bytes, got %d)", fdotconfig.CredMgrEnvVar, len(key))
			return
		}

		encryptionKey = key
	})

	return encryptionKey, keyInitError
}

// getCredFilePath returns the path to the encrypted credentials file
func getCredFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	credDir := filepath.Join(homeDir, ".local", "share", "fdot")
	return filepath.Join(credDir, "credentials.enc"), nil
}

// ensureCredDir creates the credential directory if it doesn't exist
func ensureCredDir() error {
	credFile, err := getCredFilePath()
	if err != nil {
		return err
	}

	credDir := filepath.Dir(credFile)
	if err := os.MkdirAll(credDir, 0700); err != nil {
		return fmt.Errorf("failed to create credential directory: %w", err)
	}

	return nil
}

// loadCredentials reads and decrypts the credentials file
func loadCredentials() (map[string][]byte, error) {
	credFile, err := getCredFilePath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty map
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return make(map[string][]byte), nil
	}

	// Read encrypted file
	encryptedData, err := os.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Get encryption key
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := decryptAESGCM(encryptedData, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Unmarshal JSON
	var creds map[string][]byte
	if err := json.Unmarshal(plaintext, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return creds, nil
}

// saveCredentials encrypts and writes the credentials file
func saveCredentials(creds map[string][]byte) error {
	if err := ensureCredDir(); err != nil {
		return err
	}

	credFile, err := getCredFilePath()
	if err != nil {
		return err
	}

	// Marshal to JSON
	plaintext, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Get encryption key
	key, err := getEncryptionKey()
	if err != nil {
		return err
	}

	// Encrypt
	encrypted, err := encryptAESGCM(plaintext, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(credFile, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// encryptAESGCM encrypts data using AES-256-GCM
func encryptAESGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and append nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAESGCM decrypts data using AES-256-GCM
func decryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// getCache returns the credential cache, initializing it if needed
func getCache() (map[string][]byte, error) {
	var initErr error
	credCacheInit.Do(func() {
		var err error
		credCache, err = loadCredentials()
		if err != nil {
			initErr = err
			return
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return credCache, nil
}

// readCredential reads a credential from the cache
func readCredential(name string) ([]byte, error) {
	cache, err := getCache()
	if err != nil {
		return nil, err
	}

	credCacheMutex.RLock()
	defer credCacheMutex.RUnlock()

	value, exists := cache[name]
	if !exists {
		return nil, fmt.Errorf("credential %q not found", name)
	}

	return value, nil
}

// writeCredential writes a credential to cache and persists to disk
func writeCredential(name string, value []byte) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	credCacheMutex.Lock()
	cache[name] = value
	credCacheMutex.Unlock()

	// Save to disk
	credCacheMutex.RLock()
	cacheCopy := make(map[string][]byte, len(cache))
	for k, v := range cache {
		cacheCopy[k] = v
	}
	credCacheMutex.RUnlock()

	return saveCredentials(cacheCopy)
}

// deleteCredential removes a credential from cache and persists to disk
func deleteCredential(name string) error {
	cache, err := getCache()
	if err != nil {
		return err
	}

	credCacheMutex.Lock()
	if _, exists := cache[name]; !exists {
		credCacheMutex.Unlock()
		return fmt.Errorf("credential %q not found", name)
	}
	delete(cache, name)
	credCacheMutex.Unlock()

	// Save to disk
	credCacheMutex.RLock()
	cacheCopy := make(map[string][]byte, len(cache))
	maps.Copy(cacheCopy, cache)
	credCacheMutex.RUnlock()

	return saveCredentials(cacheCopy)
}

// listCredentials returns all credential names
func listCredentials() ([]string, error) {
	cache, err := getCache()
	if err != nil {
		return nil, err
	}

	credCacheMutex.RLock()
	defer credCacheMutex.RUnlock()

	names := make([]string, 0, len(cache))
	for name := range cache {
		names = append(names, name)
	}

	return names, nil
}
