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

This implementation uses the Linux kernel keyring system with careful attention to its permission model.

**Permission Model:**
- Keys have 4 permission levels: possessor, owner, group, other
- Default permissions: `possessor=alswrv` (full access), `owner=v` (view only)
- **Important**: Matching owner UID is NOT enough to read keys!
- You need "possession" - granted when key is reachable from your session keyring

**How It Works:**
1. Credentials are stored in the user keyring (`@u`) for persistence
2. User keyring is linked into the session keyring (`@s`) via `KEYCTL_LINK`
3. This linking grants "possession", allowing `request_key()` to find and read keys
4. Without the link, even the owner UID cannot read the key content

**References:**
- Man pages: `keyctl(2)`, `keyrings(7)`, `add_key(2)`, `request_key(2)`
- Stack Overflow: https://stackoverflow.com/a/79389296
- Kernel docs: https://www.kernel.org/doc/html/latest/security/keys/core.html

**Debugging:**
- View your keyrings: `keyctl show`
- View a specific key: `keyctl describe <keyid>`
- List user keyring: `keyctl list @u`
- Check if linked: `keyctl show` should show `_uid.XXX` under `_ses`

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