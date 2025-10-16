# Linux Kernel Keyring Implementation - ARCHIVED

This directory contains the original Linux kernel keyring implementation of credmgr.

## Why Archived?

The kernel keyring implementation was replaced with AES-encrypted file storage because:
- **Kernel keyrings do NOT persist across reboots** (RAM-only, even with CONFIG_PERSISTENT_KEYRINGS)
- Windows Credential Manager DOES persist across reboots
- Platform behavior mismatch was unacceptable for credential storage

## What This Implementation Provided

✅ **Excellent security model:**
- Kernel-level protection
- Process isolation
- "Possession" permission model preventing casual access

✅ **Well-researched and documented:**
- Deep understanding of keyring permission model (possessor vs owner)
- Working demonstration code (test_possession.c)
- Comprehensive documentation of syscalls and behavior

❌ **Critical limitation:**
- All credentials lost on reboot
- Unsuitable for persistent credential storage

## Implementation Details

See the backup files:
- `credmgr_linux_keyring.go.bak` - Full implementation with syscalls
- `README.md` - Original comprehensive documentation
- `test_possession.c` - C demonstration program
- `Makefile` - Build system for demo

## Key Learnings

1. Linux kernel keyrings use "possession" (not just ownership) for access
2. Must link user keyring (@u) to session keyring (@s) via KEYCTL_LINK
3. Persistent keyrings expire after 3 days and are still RAM-only
4. Source verified in Linux kernel: security/keys/persistent.c

## Date Archived

October 16, 2025

## Replacement

New implementation uses AES-256-GCM encrypted file storage with key from
FDOT_CREDENTIAL_KEY environment variable. See `credmgr_linux.go` for details.
