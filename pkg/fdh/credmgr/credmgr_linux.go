//go:build linux

// Package credmgr provides credential management on Linux using AES-encrypted file storage.
//
// # Storage Architecture
//
// Credentials are stored in an AES-256-GCM encrypted file:
//   - Location: ~/.fdot/credentials.enc (or custom path)
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
package credmgr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/nzions/fdot/pkg/fdh"
	"github.com/nzions/fdot/pkg/fdotconfig"
)

// linuxCredManager implements CredManager for Linux using AES-encrypted file storage
type linuxCredManager struct {
	credFilePath string

	// In-memory cache of decrypted credentials
	credCache      map[string][]byte
	credCacheMutex sync.RWMutex
	credCacheInit  sync.Once

	// Cached encryption key
	encryptionKey []byte
	keyInitOnce   sync.Once
	keyInitError  error
}

// newCredManager creates a new CredManager for Linux
func newCredManager(path string) (CredManager, error) {
	if path == "" {
		// Use default path
		return defaultCredManager()
	}

	// Use specified path
	return &linuxCredManager{
		credFilePath: path,
		credCache:    make(map[string][]byte),
	}, nil
}

// defaultCredManager returns the default CredManager for Linux
func defaultCredManager() (CredManager, error) {
	// Get default path
	hd, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	defaultPath := filepath.Join(hd, ".local/credmgr", "credentials.enc")

	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(defaultPath)
	if err := fdh.CheckCreateDir(parentDir); err != nil {
		return nil, fmt.Errorf("failed to create credential directory: %w", err)
	}

	return &linuxCredManager{
		credFilePath: defaultPath,
		credCache:    make(map[string][]byte),
	}, nil
}

// getEncryptionKey loads and validates the encryption key from environment variable
func (cm *linuxCredManager) getEncryptionKey() ([]byte, error) {
	cm.keyInitOnce.Do(func() {
		keyHex := os.Getenv(fdotconfig.CredMgrEnvVarKey)
		if keyHex == "" {
			cm.keyInitError = fmt.Errorf("%s environment variable not set", fdotconfig.CredMgrEnvVarKey)
			return
		}

		key, err := hex.DecodeString(keyHex)
		if err != nil {
			cm.keyInitError = fmt.Errorf("invalid %s format (expected 64 hex chars): %w", fdotconfig.CredMgrEnvVarKey, err)
			return
		}

		if len(key) != 32 {
			cm.keyInitError = fmt.Errorf("invalid %s length (expected 32 bytes, got %d)", fdotconfig.CredMgrEnvVarKey, len(key))
			return
		}

		cm.encryptionKey = key
	})

	return cm.encryptionKey, cm.keyInitError
}

// loadCredentials reads and decrypts the credentials file
func (cm *linuxCredManager) loadCredentials() (map[string][]byte, error) {
	// If file doesn't exist, return empty map
	if _, err := os.Stat(cm.credFilePath); os.IsNotExist(err) {
		return make(map[string][]byte), nil
	}

	// Read encrypted file
	encrypted, err := os.ReadFile(cm.credFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Get encryption key
	key, err := cm.getEncryptionKey()
	if err != nil {
		return nil, err
	}

	// Decrypt
	plaintext, err := cm.decryptAESGCM(encrypted, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Unmarshal JSON
	var creds map[string][]byte
	if err := json.Unmarshal(plaintext, &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return creds, nil
}

// saveCredentials encrypts and writes the credentials file
func (cm *linuxCredManager) saveCredentials(creds map[string][]byte) error {
	// Ensure directory exists
	if err := fdh.CheckCreateDir(filepath.Dir(cm.credFilePath)); err != nil {
		return err
	}

	// Marshal to JSON
	plaintext, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Get encryption key
	key, err := cm.getEncryptionKey()
	if err != nil {
		return err
	}

	// Encrypt
	encrypted, err := cm.encryptAESGCM(plaintext, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(cm.credFilePath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// getCache returns the in-memory credential cache, loading it if necessary
func (cm *linuxCredManager) getCache() (map[string][]byte, error) {
	var loadErr error
	cm.credCacheInit.Do(func() {
		var err error
		cm.credCache, err = cm.loadCredentials()
		if err != nil {
			loadErr = err
			return
		}
	})

	if loadErr != nil {
		return nil, loadErr
	}

	return cm.credCache, nil
}

// encryptAESGCM encrypts plaintext using AES-256-GCM
func (cm *linuxCredManager) encryptAESGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAESGCM decrypts ciphertext using AES-256-GCM
func (cm *linuxCredManager) decryptAESGCM(ciphertext, key []byte) ([]byte, error) {
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
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Implementation of CredManager interface methods

// Read retrieves raw credential bytes by name.
func (cm *linuxCredManager) Read(name string) ([]byte, error) {
	cache, err := cm.getCache()
	if err != nil {
		return nil, err
	}

	cm.credCacheMutex.RLock()
	defer cm.credCacheMutex.RUnlock()

	data, exists := cache[name]
	if !exists {
		return nil, fmt.Errorf("credential %q %w", name, ErrNotFound)
	}

	return data, nil
}

// Write stores raw credential bytes with the given name.
func (cm *linuxCredManager) Write(name string, data []byte) error {
	cache, err := cm.getCache()
	if err != nil {
		return err
	}

	cm.credCacheMutex.Lock()
	cache[name] = data
	cm.credCacheMutex.Unlock()

	// Save to disk
	cm.credCacheMutex.RLock()
	cacheCopy := make(map[string][]byte, len(cache))
	maps.Copy(cacheCopy, cache)
	cm.credCacheMutex.RUnlock()

	return cm.saveCredentials(cacheCopy)
}

// ReadKey retrieves a credential key as a string.
func (cm *linuxCredManager) ReadKey(name string) (string, error) {
	data, err := cm.Read(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteKey stores a string credential key.
func (cm *linuxCredManager) WriteKey(name, key string) error {
	return cm.Write(name, []byte(key))
}

// ReadUserCred retrieves a username/password credential.
func (cm *linuxCredManager) ReadUserCred(name string) (UserCred, error) {
	data, err := cm.Read(name)
	if err != nil {
		return nil, err
	}
	return unmarshalUnPw(data)
}

// WriteUserCred stores a username/password credential.
func (cm *linuxCredManager) WriteUserCred(name string, cred UserCred) error {
	// Type assert to access marshal method
	if uc, ok := cred.(*obfuscatedUserCred); ok {
		return cm.Write(name, uc.marshal())
	}
	// Fallback: reconstruct from interface
	reconstructed := newObfuscatedUserCred(cred.Username(), cred.Password())
	return cm.Write(name, reconstructed.marshal())
}

// Delete removes a credential by name.
func (cm *linuxCredManager) Delete(name string) error {
	cache, err := cm.getCache()
	if err != nil {
		return err
	}

	cm.credCacheMutex.Lock()
	if _, exists := cache[name]; !exists {
		cm.credCacheMutex.Unlock()
		return fmt.Errorf("credential %q %w", name, ErrNotFound)
	}
	delete(cache, name)
	cm.credCacheMutex.Unlock()

	// Save to disk
	cm.credCacheMutex.RLock()
	cacheCopy := make(map[string][]byte, len(cache))
	maps.Copy(cacheCopy, cache)
	cm.credCacheMutex.RUnlock()

	return cm.saveCredentials(cacheCopy)
}

// DeleteDB removes the entire credential database.
func (cm *linuxCredManager) DeleteDB() error {
	// Clear the in-memory cache first
	cm.credCacheMutex.Lock()
	cm.credCache = make(map[string][]byte)
	cm.credCacheMutex.Unlock()

	// Remove the encrypted file if it exists
	if _, err := os.Stat(cm.credFilePath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to stat credentials file: %w", err)
	}

	if err := os.Remove(cm.credFilePath); err != nil {
		return fmt.Errorf("failed to delete credentials database: %w", err)
	}

	return nil
}

// List returns all credential names.
func (cm *linuxCredManager) List() ([]string, error) {
	cache, err := cm.getCache()
	if err != nil {
		return nil, err
	}

	cm.credCacheMutex.RLock()
	defer cm.credCacheMutex.RUnlock()

	names := make([]string, 0, len(cache))
	for name := range cache {
		names = append(names, name)
	}

	return names, nil
}
