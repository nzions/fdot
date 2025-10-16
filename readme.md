# FDOT Code

All code in this repository is the property of FDOT (Florida Department of Transport).

## Projects

### Credential Manager

A cross-platform credential management library and CLI tool for securely storing and retrieving sensitive data.

#### Features

- **Cross-platform**: Works on Windows (Credential Manager) and Linux (kernel keyring)
- **Secure storage**: Uses OS-native secure storage mechanisms
- **Simple API**: Clean Go interface with minimal dependencies
- **CLI tool**: Command-line interface for credential management
- **No forced prefixing**: Store credentials with clean names

#### Installation

```bash
go build -o credmgr ./cmd/credmgr
```

#### Usage

**CLI Tool:**
```bash
# Store a credential
./credmgr set myapp-token "secret123"

# Retrieve a credential
./credmgr get myapp-token

# List all credentials
./credmgr list

# Delete a credential
./credmgr del myapp-token
```

**Go Library:**
```go
import "github.com/nzions/fdot/pkg/fdh/credmgr"

// Store a credential
err := credmgr.WriteString("api-key", "secret-value")

// Retrieve a credential
value, err := credmgr.ReadString("api-key")

// Delete a credential
err := credmgr.Delete("api-key")

// List all credentials
names, err := credmgr.List()
```

#### Architecture

- **Windows**: Uses Win32 Credential Manager API via syscalls
- **Linux**: Uses kernel keyring API via syscalls
- **Build tags**: Platform-specific implementations with unified interface

#### Build

```bash
# Linux build
go build -o credmgr ./cmd/credmgr

# Windows cross-compile
GOOS=windows go build -o credmgr.exe ./cmd/credmgr
```