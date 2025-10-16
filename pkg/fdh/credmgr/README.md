# credmgr - Cross-Platform Credential Manager

Unified credential management package that works across Windows and Linux platforms.

## Architecture

```
pkg/fdh/credmgr/
├── credmgr.go          # Public API (exported functions)
├── credmgr_windows.go  # Windows implementation (build tag: windows)
├── credmgr_linux.go    # Linux implementation (build tag: linux)
└── README.md           # This file
```

## Design Principles

- **DRY**: Single interface, platform-specific implementations
- **KISS**: Simple 5-function API
- **YAGNI**: Only essential credential operations
- **Build Tags**: Compile-time platform selection

## API

```go
// Core operations
func Read(name string) ([]byte, error)
func Write(name string, data []byte) error
func Delete(name string) error
func List() ([]string, error)

// String convenience functions
func ReadString(name string) (string, error)
func WriteString(name, value string) error
```

## Platform Support

### Windows
- **Backend**: Windows Credential Manager API
- **Storage**: `CredRead`, `CredWrite`, `CredDelete`, `CredEnumerate`
- **Persistence**: Local machine scope
- **Security**: Windows built-in credential encryption

### Linux  
- **Backend**: Linux kernel keyring system
- **Storage**: `add_key`, `request_key`, `keyctl` syscalls
- **Persistence**: User keyring (persistent across sessions)
- **Security**: Kernel-level isolation and permissions

## Usage

```go
import "github.com/nzions/fdot/pkg/fdh/credmgr"

// Store credential
err := credmgr.WriteString("app-token", "secret123")

// Retrieve credential
token, err := credmgr.ReadString("app-token")

// Delete credential
err := credmgr.Delete("app-token")

// List all credentials
names, err := credmgr.List()
```

## Error Handling

- `credmgr.ErrNotFound`: Credential does not exist
- `credmgr.ErrNotSupported`: Platform not supported (should not happen with current build tags)

## Implementation Notes

### Linux Kernel Keyring
- Keys stored with prefix `fdot-cred:` to avoid naming conflicts
- Uses user keyring (`KEY_SPEC_USER_KEYRING`) for persistence
- Requires appropriate user permissions (should work for any user)
- Data visible in `/proc/keys` (for debugging)

### Windows Credential Manager
- Uses `CRED_TYPE_GENERIC` for application credentials  
- `CRED_PERSIST_LOCAL_MACHINE` for machine-wide storage
- Integrates with Windows Credential Manager GUI
- 2560 byte limit per credential (Windows API limitation)

## Cross-Platform Compatibility

The same code compiles and works on both platforms:
```bash
# Linux build
GOOS=linux go build

# Windows build  
GOOS=windows go build
```

Both implementations provide identical behavior and error handling for seamless cross-platform deployment.