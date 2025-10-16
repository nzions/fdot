// Demonstrates the Linux kernel keyring "possession" permission model
// Compile: make
// Run: make run  OR  ./test_possession

#define _GNU_SOURCE
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <unistd.h>
#include <stdint.h>
#include <stdarg.h>
#include <sys/syscall.h>

// keyctl operations
#define KEYCTL_READ 11
#define KEYCTL_LINK 8

// Special keyring IDs
#define KEY_SPEC_SESSION_KEYRING -3
#define KEY_SPEC_USER_KEYRING -4

typedef int32_t key_serial_t;

static key_serial_t add_key(const char *type, const char *description,
                            const void *payload, size_t plen, key_serial_t keyring) {
    return (key_serial_t)syscall(SYS_add_key, type, description, payload, plen, keyring);
}

static key_serial_t request_key(const char *type, const char *description,
                                const char *callout_info, key_serial_t dest_keyring) {
    return (key_serial_t)syscall(SYS_request_key, type, description, callout_info, dest_keyring);
}

static long keyctl_call(int operation, ...) {
    va_list args;
    va_start(args, operation);
    unsigned long arg2 = va_arg(args, unsigned long);
    unsigned long arg3 = va_arg(args, unsigned long);
    unsigned long arg4 = va_arg(args, unsigned long);
    va_end(args);
    return syscall(SYS_keyctl, operation, arg2, arg3, arg4);
}

int main() {
    printf("=== Linux Kernel Keyring Possession Demo ===\n\n");
    
    // Test 1: Without linking (demonstrates the problem)
    printf("TEST 1: Adding key WITHOUT linking user keyring to session keyring\n");
    printf("--------------------------------------------------------------------\n");
    
    key_serial_t key1 = add_key("user", "test-no-link", "secret1", 7, KEY_SPEC_USER_KEYRING);
    if (key1 < 0) {
        printf("❌ add_key failed: %s\n", strerror(errno));
        return 1;
    }
    printf("✅ Key added to @u with ID: %d\n", key1);
    
    // Try to read it
    char buffer[256];
    long result = keyctl_call(KEYCTL_READ, key1, buffer, sizeof(buffer));
    if (result < 0) {
        printf("❌ Reading key: FAILED - %s\n", strerror(errno));
        printf("   Reason: Owner UID only grants 'view' permission, not 'read'\n");
        printf("   You need 'possession' permission to read the key content\n");
    } else {
        printf("✅ Reading key: SUCCESS (data: %.*s)\n", (int)result, buffer);
    }
    
    // Try to find it with request_key
    key_serial_t found1 = request_key("user", "test-no-link", NULL, KEY_SPEC_USER_KEYRING);
    if (found1 < 0) {
        printf("❌ request_key(@u): FAILED - %s\n", strerror(errno));
        printf("   Reason: Key is not 'possessed' (not reachable from @s)\n\n");
    } else {
        printf("✅ request_key(@u): SUCCESS (found key: %d)\n\n", found1);
    }
    
    // Test 2: With linking (demonstrates the solution)
    printf("TEST 2: Adding key AFTER linking user keyring to session keyring\n");
    printf("------------------------------------------------------------------\n");
    
    // Link @u into @s - this is the magic!
    long link_result = keyctl_call(KEYCTL_LINK, KEY_SPEC_USER_KEYRING, KEY_SPEC_SESSION_KEYRING);
    if (link_result < 0) {
        printf("❌ keyctl_call(KEYCTL_LINK) failed: %s\n", strerror(errno));
        return 1;
    }
    printf("✅ Linked @u into @s (grants 'possession' permission)\n");
    
    key_serial_t key2 = add_key("user", "test-with-link", "secret2", 7, KEY_SPEC_USER_KEYRING);
    if (key2 < 0) {
        printf("❌ add_key failed: %s\n", strerror(errno));
        return 1;
    }
    printf("✅ Key added to @u with ID: %d\n", key2);
    
    // Try to read it (should work now!)
    result = keyctl_call(KEYCTL_READ, key2, buffer, sizeof(buffer));
    if (result < 0) {
        printf("❌ Reading key: FAILED - %s\n", strerror(errno));
    } else {
        printf("✅ Reading key: SUCCESS (data: %.*s)\n", (int)result, buffer);
        printf("   Reason: Linking @u into @s granted 'possession' permission\n");
    }
    
    // Try to find it with request_key (should work now!)
    key_serial_t found2 = request_key("user", "test-with-link", NULL, 0);
    if (found2 < 0) {
        printf("❌ request_key(default): FAILED - %s\n", strerror(errno));
    } else {
        printf("✅ request_key(default): SUCCESS (found key: %d)\n", found2);
        printf("   Reason: Kernel searches @s → @u and finds the key\n\n");
    }
    
    // Bonus: Try to read the first key now that @u is linked
    printf("BONUS: Can we read the first key now that @u is linked?\n");
    printf("--------------------------------------------------------\n");
    result = keyctl_call(KEYCTL_READ, key1, buffer, sizeof(buffer));
    if (result < 0) {
        printf("❌ Reading first key: STILL FAILS - %s\n", strerror(errno));
        printf("   The linking doesn't retroactively grant possession\n");
    } else {
        printf("✅ Reading first key: SUCCESS (data: %.*s)\n", (int)result, buffer);
        printf("   The linking DOES work retroactively! All keys in @u are now possessed\n");
    }
    
    // And request_key should work for the first key now
    found1 = request_key("user", "test-no-link", NULL, 0);
    if (found1 < 0) {
        printf("❌ request_key(default) for first key: FAILED - %s\n\n", strerror(errno));
    } else {
        printf("✅ request_key(default) for first key: SUCCESS (found key: %d)\n", found1);
        printf("   Both KEYCTL_READ and request_key work after linking\n\n");
    }
    
    printf("=== Summary ===\n");
    printf("1. Default permissions: possessor=alswrv, owner=v (view only)\n");
    printf("2. Matching owner UID is NOT enough - you need 'possession'\n");
    printf("3. Possession is granted by linking @u into @s\n");
    printf("4. Use request_key() to find keys, not direct key IDs\n");
    printf("5. This is standard Linux kernel keyring behavior, not a bug\n\n");
    printf("Reference: https://stackoverflow.com/a/79389296\n");
    
    return 0;
}
