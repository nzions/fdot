# Documentation

## Linux Kernel Keyring

Detailed documentation about the Linux kernel keyring "possession" permission model and why credential manager needs to link the user keyring to the session keyring.

**Location:** [`linux-kernel-keyring/`](linux-kernel-keyring/)

**Contents:**
- `README.md` - Comprehensive technical documentation
- `test_possession.c` - Demonstration program showing with/without linking behavior
- `Makefile` - Build and run the demonstration

**Quick Start:**
```bash
cd docs/linux-kernel-keyring
make run
```

This demonstrates why matching the owner UID is not sufficient to read keys - you need "possession" permission, which is granted by linking the user keyring (@u) into the session keyring (@s).

**Key Reference:** https://stackoverflow.com/a/79389296
