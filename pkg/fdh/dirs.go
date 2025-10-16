package fdh

import (
	"errors"
	"os"
)

var (
	ErrNotDir      = errors.New("not a directory")
	ErrNotWritable = errors.New("not writable")
)

// CheckCreateDir checks if a directory exists and is writable, creating it if it doesn't exist.
// Returns an error if the path exists but is not a directory or is not writable.
func CheckCreateDir(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err == nil {
		// Directory exists, make sure it's actually a directory
		if !info.IsDir() {
			return ErrNotDir
		}

		// Check if directory is writable (owner has write permission)
		if info.Mode().Perm()&0200 == 0 {
			return ErrNotWritable
		}

		return nil
	}

	// Directory doesn't exist, create it with proper permissions
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}
