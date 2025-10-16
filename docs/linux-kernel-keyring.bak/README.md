# Linux Kernel Keyring Implementation Notes

## The Problem We Solved

When implementing credential storage using Linux kernel keyrings, we discovered that keys added to the user keyring (`@u`) could not be read, even by the process that created them, returning `EPERM` (Permission denied).

## Root Cause: The "Possession" Permission Model

Linux kernel keyrings use a **unique 4-level permission model**:

1. **possessor** - Has the key reachable from their session keyring
2. **owner** - Matches the key's owner UID  
3. **group** - Matches the key's group GID
4. **other** - Everyone else

### Default Key Permissions

When you add a key, default permissions are typically:
```
alswrv-----v------------
```

Breaking this down:
- **possessor**: `alswrv` (all permissions: alter, link, search, write, read, view)
- **owner**: `-----v` (view only - can see it exists but **cannot read content**)
- **group**: `------` (no access)
- **other**: `------` (no access)

### The Critical Insight

**Matching the owner UID is NOT sufficient to read a key!**

You need **"possession"**, which is granted when a key is **reachable from your process's session keyring** (`@s`).

## How Possession Works

When you run `keyctl show` (showing your session keyring), you'll see:

```bash
$ keyctl show
Session Keyring
1046546095 --alswrv   1000   100  keyring: _ses
 348545926 --alswrv   1000 65534   \_ keyring: _uid.1000
 987654321 --alswrv   1000  1000       \_ user: my-credential
```

The `_uid.1000` entry is a **link** from the session keyring to your user keyring. This link is what grants you "possession" of keys stored in `@u`.

The kernel's search path goes: `@s` → `@u` (if linked), and if the link exists, you "possess" the keys.

## The Solution: Linking Keyrings

Our implementation ensures the user keyring is linked to the session keyring:

```go
func ensureUserKeyringLinked() {
    syscall.Syscall(sysKeyctl,
        keyctlLink,           // KEYCTL_LINK = 8
        keySpecUserKeyring,   // @u = -4
        keySpecSessionKeyring) // @s = -3
    // Errors ignored - if already linked or fails, that's ok
}
```

This grants us "possession" permission, allowing us to:
1. Store keys in `@u` for persistence
2. Read keys via `request_key()` which searches `@s` → `@u`
3. Access our own credentials without elevated privileges

## Why This Approach

### Benefits of User Keyring (@u)
- **Persistence**: Keys survive across terminal sessions
- **User-scoped**: Isolated from other users
- **Kernel-protected**: Cannot be read by other processes without possession

### Benefits of Session Linking
- **Possession**: Grants full access (read/write/delete)
- **Discoverability**: `request_key()` can find keys via search path
- **Standard pattern**: PAM modules (like `pam_keyinit`) do this automatically on login

## Common Misconceptions Debunked

❌ **"This is a WSL2 bug"**  
✅ This is standard Linux kernel keyring behavior on all systems

❌ **"The owner should be able to read their own keys"**  
✅ Owner UID only grants "view" permission by default, not "read"

❌ **"Keys in @u should be directly readable"**  
✅ You need the key to be "possessed" (reachable from @s) to read it

❌ **"We need CONFIG_PERSISTENT_KEYRINGS"**  
✅ That's for `@p` (persistent keyring), not required for `@u` (user keyring)

## Testing the Behavior

### Without Linking (Default)
```c
// Add key to user keyring
key_serial_t key = add_key("user", "test", "data", 4, KEY_SPEC_USER_KEYRING);

// Try to read - FAILS with EPERM!
keyctl(KEYCTL_READ, key, buffer, sizeof(buffer));  // Permission denied

// Try to find with request_key - FAILS with ENOKEY!
key_serial_t found = request_key("user", "test", NULL, KEY_SPEC_USER_KEYRING);
```

### With Linking (Our Solution)
```c
// Link user keyring into session keyring
keyctl(KEYCTL_LINK, KEY_SPEC_USER_KEYRING, KEY_SPEC_SESSION_KEYRING);

// Add key to user keyring
key_serial_t key = add_key("user", "test", "data", 4, KEY_SPEC_USER_KEYRING);

// Now reading works!
keyctl(KEYCTL_READ, key, buffer, sizeof(buffer));  // SUCCESS

// And request_key finds it!
key_serial_t found = request_key("user", "test", NULL, 0);  // SUCCESS (searches @s → @u)
```

## References

### Documentation
- Man pages: `keyctl(2)`, `keyrings(7)`, `add_key(2)`, `request_key(2)`
- Kernel docs: https://www.kernel.org/doc/html/latest/security/keys/core.html

### Authoritative Explanation
- Stack Overflow: https://stackoverflow.com/a/79389296
  - Detailed explanation by grawity (Linux expert)
  - Confirms this is standard behavior
  - Shows PAM module handling

### Our Implementation
- `pkg/fdh/credmgr/credmgr_linux.go` - Full implementation with detailed comments
- `pkg/fdh/credmgr/README.md` - User documentation

## Key Takeaways

1. **Kernel keyrings use "possession" as the primary permission mechanism**
2. **Linking @u into @s grants possession** - this is not a workaround, it's the correct approach
3. **PAM modules do this automatically on login** - we're replicating standard behavior
4. **This works on all Linux systems**, not just WSL2
5. **Our implementation follows Linux kernel keyring best practices**

## Debugging Commands

```bash
# View your session keyring
keyctl show

# View your user keyring
keyctl show @u

# List keys in user keyring
keyctl list @u

# Describe a specific key (shows permissions)
keyctl describe <keyid>

# Manually link keyrings (what our code does)
keyctl link @u @s
```

---

**Author**: GitHub Copilot  
**Date**: October 16, 2025  
**Project**: fdot - Florida Department of Transport credential manager
