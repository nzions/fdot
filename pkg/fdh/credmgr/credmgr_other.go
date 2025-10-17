//go:build !windows && !linux

package credmgr

// otherCredManager implements CredManager for unsupported platforms
type otherCredManager struct{}

// newCredManager creates a new CredManager for other platforms (returns not supported)
func newCredManager(path string) (CredManager, error) {
	return &otherCredManager{}, nil
}

// defaultCredManager returns the default CredManager for other platforms (returns not supported)
func defaultCredManager() (CredManager, error) {
	return &otherCredManager{}, nil
}

// All methods return ErrNotSupported for unsupported platforms

func (om *otherCredManager) Read(name string) ([]byte, error) {
	return nil, ErrNotSupported
}

func (om *otherCredManager) Write(name string, data []byte) error {
	return ErrNotSupported
}

func (om *otherCredManager) ReadKey(name string) (string, error) {
	return "", ErrNotSupported
}

func (om *otherCredManager) WriteKey(name, key string) error {
	return ErrNotSupported
}

func (om *otherCredManager) ReadUserCred(name string) (UserCred, error) {
	return nil, ErrNotSupported
}

func (om *otherCredManager) WriteUserCred(name string, cred UserCred) error {
	return ErrNotSupported
}

func (om *otherCredManager) Delete(name string) error {
	return ErrNotSupported
}

func (om *otherCredManager) DeleteDB() error {
	return ErrNotSupported
}

func (om *otherCredManager) List() ([]string, error) {
	return nil, ErrNotSupported
}
