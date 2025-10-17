package fdotconfig

import (
	"os"
	"path/filepath"
)

const (
	FDOTDir           = ".fdot"
	BigKeySecretName  = "fdh-user-bigkey"
	SSHCredSecretName = "fdh-user-ssh-creds"
	CredMgrEnvVarKey  = "CREDMGR_KEY" // linux only
	CredMgrEnvVarPath = "CREDMGR_DIR" // linux only
)

// PathProvider defines an interface for providing credential file paths.
// This allows different implementations while avoiding import cycles.
type PathProvider interface {
	CredFilePath() string
}

// Default implementation that can be used when no PathProvider is available
var defaultPathProvider PathProvider

// SetPathProvider sets the default path provider for credential operations
func SetPathProvider(provider PathProvider) {
	defaultPathProvider = provider
}

// GetCredFilePath returns the credential file path using the registered provider
// or falls back to a basic implementation if no provider is set
func GetCredFilePath() (string, error) {
	if defaultPathProvider != nil {
		return defaultPathProvider.CredFilePath(), nil
	}

	// Fallback implementation
	credDir := os.Getenv(CredMgrEnvVarPath)
	if credDir != "" {
		return filepath.Join(credDir, "credentials.enc"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, FDOTDir, "credentials.enc"), nil
}
