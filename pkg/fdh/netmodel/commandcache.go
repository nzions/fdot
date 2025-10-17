package netmodel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CommandCache manages reading and writing command outputs to disk
type CommandCache struct {
	config *CacheConfig
}

// NewCommandCache creates a new command cache manager
func NewCommandCache(config *CacheConfig) *CommandCache {
	if config == nil {
		config = DefaultCacheConfig()
	}
	return &CommandCache{
		config: config,
	}
}

// GetCachedOutput attempts to read cached output for a command
// Returns the cached output and true if found and not expired, or empty string and false otherwise
func (c *CommandCache) GetCachedOutput(deviceIP, command string) (string, bool) {
	if !c.config.Enabled {
		return "", false
	}

	filePath := c.getCacheFilePath(deviceIP, command)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return "", false // File doesn't exist
	}

	// Check if file is expired based on TTL
	if time.Since(info.ModTime()) > c.config.TTL {
		return "", false // Cache expired
	}

	// Read and return cached content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", false
	}

	return string(content), true
}

// SaveOutput saves command output to cache file
func (c *CommandCache) SaveOutput(deviceIP, command, output string) error {
	if !c.config.Enabled {
		return nil // Caching disabled, nothing to save
	}

	filePath := c.getCacheFilePath(deviceIP, command)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write output to file
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getCacheFilePath generates a consistent file path for a command
// Format: <baseDir>/<deviceIP>/<command_hash>.txt
func (c *CommandCache) getCacheFilePath(deviceIP, command string) string {
	baseDir := c.config.BaseDir
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "fdot-cache")
	}

	// Sanitize device IP (replace colons and dots with underscores for IPv6/IPv4)
	sanitizedIP := strings.ReplaceAll(deviceIP, ":", "_")
	sanitizedIP = strings.ReplaceAll(sanitizedIP, ".", "_")

	// Create hash of command for filename (handles special chars and length)
	hash := sha256.Sum256([]byte(command))
	commandHash := hex.EncodeToString(hash[:])[:16] // Use first 16 chars of hash

	// Create filename with command prefix for readability
	commandPrefix := sanitizeCommandForFilename(command)
	filename := fmt.Sprintf("%s_%s.txt", commandPrefix, commandHash)

	return filepath.Join(baseDir, sanitizedIP, filename)
}

// sanitizeCommandForFilename creates a safe filename prefix from command
// Takes first few words of command and removes special characters
func sanitizeCommandForFilename(command string) string {
	// Take first 30 chars, replace spaces with underscores
	prefix := command
	if len(prefix) > 30 {
		prefix = prefix[:30]
	}

	// Replace spaces and special chars
	prefix = strings.ReplaceAll(prefix, " ", "_")
	prefix = strings.ReplaceAll(prefix, "/", "_")
	prefix = strings.ReplaceAll(prefix, "|", "_")
	prefix = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, prefix)

	return prefix
}

// ClearCache removes all cached files for a device
func (c *CommandCache) ClearCache(deviceIP string) error {
	if !c.config.Enabled {
		return nil
	}

	baseDir := c.config.BaseDir
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "fdot-cache")
	}

	sanitizedIP := strings.ReplaceAll(deviceIP, ":", "_")
	sanitizedIP = strings.ReplaceAll(sanitizedIP, ".", "_")

	deviceDir := filepath.Join(baseDir, sanitizedIP)

	// Remove the entire device directory
	if err := os.RemoveAll(deviceDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}
