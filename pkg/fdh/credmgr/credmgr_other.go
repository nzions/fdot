//go:build !windows && !linux

package credmgr

// readCredential is not implemented on this platform
func readCredential(name string) ([]byte, error) {
	return nil, ErrNotSupported
}

// writeCredential is not implemented on this platform
func writeCredential(name string, data []byte) error {
	return ErrNotSupported
}

// deleteCredential is not implemented on this platform
func deleteCredential(name string) error {
	return ErrNotSupported
}

// listCredentials is not implemented on this platform
func listCredentials() ([]string, error) {
	return nil, ErrNotSupported
}