# Migration from Kernel Keyring to File-Based Storage

**Date:** October 16, 2025

## Summary

Replaced Linux kernel keyring implementation with AES-256-GCM encrypted file storage to provide cross-reboot credential persistence.

## What Changed

### Before (Kernel Keyring)
- **Storage**: Linux kernel keyrings (RAM-only)
- **Persistence**: Session-only (lost on reboot)
- **Security**: Excellent (kernel-level protection, possession model)
- **API**: `add_key`, `request_key`, `keyctl` syscalls

### After (Encrypted File)
- **Storage**: `~/.local/share/fdot/credentials.enc`
- **Persistence**: File-based (survives reboots) ✅
- **Security**: Good (AES-256-GCM encryption, 0600 permissions)
- **API**: Standard Go file I/O + crypto/aes

## Usage

### Required Environment Variable

```bash
# Generate encryption key (do once)
openssl rand -hex 32

# Set in environment (add to ~/.bashrc or ~/.profile)
export CREDMGR_KEY="<your-64-character-hex-key>"
```

### Testing

```bash
# Build
go build -o bin/credmgr ./cmd/credmgr

# Store credential
./bin/credmgr set mykey myvalue

# Retrieve credential
./bin/credmgr get mykey

# List all credentials
./bin/credmgr list

# Delete credential
./bin/credmgr delete mykey
```

## Files Modified

1. **`pkg/fdh/credmgr/credmgr_linux.go`** - Complete rewrite
   - Removed all syscall code
   - Added AES-256-GCM encryption/decryption
   - Added file-based storage with in-memory cache
   - Added environment variable key loading

2. **`pkg/fdh/credmgr/README.md`** - Updated documentation
   - Changed Linux backend description
   - Added setup instructions for CREDMGR_KEY
   - Updated security model explanation

## Files Backed Up

1. **`pkg/fdh/credmgr/credmgr_linux_keyring.go.bak`** - Original implementation
2. **`docs/linux-kernel-keyring.bak/`** - Complete documentation archive
   - `BACKUP_README.md` - Why archived and key learnings
   - `README.md` - Original technical docs
   - `test_possession.c` - Working C demo program
   - `Makefile` - Build system

## Security Comparison

| Aspect | Kernel Keyring | Encrypted File |
|--------|---------------|----------------|
| Encryption at rest | ✅ (in RAM) | ✅ (on disk) |
| Persists across reboots | ❌ | ✅ |
| Protected from root | ❌ | ❌ |
| Protected from user | ✅ (kernel) | ⚠️ (if key exposed) |
| Cross-platform parity | ❌ (Linux only) | ✅ (matches Windows) |

## Rationale

Linux kernel keyrings are **RAM-only** by design, even with `CONFIG_PERSISTENT_KEYRINGS` enabled (verified in kernel source: `security/keys/persistent.c`). This makes them unsuitable for credential storage that needs to persist across reboots.

Windows Credential Manager **does** persist across reboots, creating a platform behavior mismatch that was unacceptable for cross-platform credential management.

The encrypted file approach:
- Matches Windows behavior (persistence)
- Provides reasonable security for development credentials
- User controls the encryption key
- Simple and maintainable

## Key Learnings Preserved

The kernel keyring investigation taught us valuable lessons about Linux security:

1. **"Possession" vs "Ownership"** in kernel keyring permission model
2. Need to link user keyring (@u) to session keyring (@s) via KEYCTL_LINK
3. Persistent keyrings expire after 3 days and are still RAM-only
4. No file I/O in kernel keys subsystem (verified via kernel source)

This knowledge is preserved in `docs/linux-kernel-keyring.bak/`.

## Testing Results

✅ **All operations working:**
- Write: `credentials.enc` created with 0600 permissions
- Read: Decryption and retrieval successful
- List: All credential names returned
- Delete: Credentials removed and file updated
- Error handling: Proper errors when `CREDMGR_KEY` not set
- File format: Encrypted data (verified with `od`)
- Persistence: File survives across sessions

## Next Steps

Users need to:
1. Generate their encryption key: `openssl rand -hex 32`
2. Add `export CREDMGR_KEY="..."` to their shell profile
3. Rebuild applications using credmgr

## Compatibility

- **Go version**: 1.24.4+
- **Linux**: Any (no kernel features required)
- **Windows**: No changes (still uses Credential Manager API)
