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
- **Backend**: AES-256-GCM encrypted file storage
- **Storage**: `~/.local/share/fdot/credentials.enc` (file permissions: 0600)
- **Encryption Key**: Environment variable `FDOT_CREDENTIAL_KEY` (64 hex chars)
- **Persistence**: File-based (survives reboots)
- **Security**: AES-256-GCM authenticated encryption

## Usage

### Linux Setup

First, generate and set your encryption key:

```bash
# Generate a random 32-byte key (do this once)
openssl rand -hex 32

# Set it in your environment (add to ~/.bashrc or ~/.profile)
export FDOT_CREDENTIAL_KEY="your-64-hex-character-key-here"
```

### Example Code

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

### Linux: AES-256-GCM Encrypted File Storage

**Storage Location:**
- File: `~/.local/share/fdot/credentials.enc`
- Permissions: `0600` (owner read/write only)
- Directory permissions: `0700`

**Encryption:**
- Algorithm: AES-256-GCM (Galois/Counter Mode)
- Key size: 256 bits (32 bytes)
- Key source: `FDOT_CREDENTIAL_KEY` environment variable
- Format: 64 hexadecimal characters
- Authenticated encryption: Protects against tampering

**Security Model:**
- ✅ Encrypted at rest
- ✅ Per-user file isolation (Unix permissions)
- ✅ Persists across reboots
- ✅ Protection against casual file viewing
- ❌ Does NOT protect against root access
- ❌ Does NOT protect if attacker has both file AND key
- ⚠️ Key management is user's responsibility

**Design Rationale:**
Previous implementation used Linux kernel keyrings, which provided excellent RAM-based
security but did NOT persist across reboots. This file-based approach matches Windows
Credential Manager behavior (persistence + reasonable security for development credentials).

For archived kernel keyring implementation, see: `docs/linux-kernel-keyring.bak/`

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