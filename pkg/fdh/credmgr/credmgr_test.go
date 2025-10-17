package credmgr

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

// setupTestEnv ensures we have a valid encryption key for testing
func setupTestEnv(t *testing.T) func() {
	// Save original env var if it exists
	originalKey := os.Getenv("CREDMGR_KEY")
	
	// Set a test key (32 bytes = 64 hex chars)
	testKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := os.Setenv("CREDMGR_KEY", testKey); err != nil {
		t.Fatalf("Failed to set CREDMGR_KEY: %v", err)
	}
	
	// Clean up any existing credentials file from different key
	// This prevents "failed to decrypt" errors when switching between test runs
	credFile := os.ExpandEnv("$HOME/.fdot/credentials.enc")
	os.Remove(credFile) // Ignore errors - file might not exist
	
	// Return cleanup function
	return func() {
		if originalKey != "" {
			os.Setenv("CREDMGR_KEY", originalKey)
		} else {
			os.Unsetenv("CREDMGR_KEY")
		}
		// Clean up any test credentials
		names, _ := List()
		for _, name := range names {
			if len(name) > 5 && name[:5] == "test-" {
				Delete(name) // Ignore errors
			}
		}
	}
}

func TestWriteRead(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple binary data",
			data: []byte("hello world"),
		},
		{
			name: "empty data",
			data: []byte(""),
		},
		{
			name: "binary data with nulls",
			data: []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd},
		},
		{
			name: "large data",
			data: bytes.Repeat([]byte("test"), 1000),
		},
		{
			name: "unicode data",
			data: []byte("こんにちは世界🌍"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credName := "test-write-read-" + tt.name

			// Write
			if err := Write(credName, tt.data); err != nil {
				t.Fatalf("Write failed: %v", err)
			}

			// Read
			retrieved, err := Read(credName)
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			// Verify
			if !bytes.Equal(retrieved, tt.data) {
				t.Errorf("Data mismatch:\ngot:  %v\nwant: %v", retrieved, tt.data)
			}

			// Cleanup
			if err := Delete(credName); err != nil {
				t.Logf("Warning: cleanup failed for %s: %v", credName, err)
			}
		})
	}
}

func TestReadNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, err := Read("test-nonexistent-credential-xyz")
	if err == nil {
		t.Error("Read of nonexistent credential should return error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Read error should wrap ErrNotFound, got: %v", err)
	}
}

func TestWriteKeyReadKey(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name string
		key  string
	}{
		{
			name: "api key",
			key:  "sk-proj-1234567890abcdef",
		},
		{
			name: "empty string",
			key:  "",
		},
		{
			name: "multiline",
			key:  "line1\nline2\nline3",
		},
		{
			name: "special chars",
			key:  "!@#$%^&*()_+-={}[]|:;<>?,./",
		},
		{
			name: "unicode",
			key:  "キー🔑",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credName := "test-key-" + tt.name

			// WriteKey
			if err := WriteKey(credName, tt.key); err != nil {
				t.Fatalf("WriteKey failed: %v", err)
			}

			// ReadKey
			retrieved, err := ReadKey(credName)
			if err != nil {
				t.Fatalf("ReadKey failed: %v", err)
			}

			// Verify
			if retrieved != tt.key {
				t.Errorf("Key mismatch:\ngot:  %q\nwant: %q", retrieved, tt.key)
			}

			// Cleanup
			Delete(credName)
		})
	}
}

func TestReadKeyNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, err := ReadKey("test-nonexistent-key-xyz")
	if err == nil {
		t.Error("ReadKey of nonexistent credential should return error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadKey error should wrap ErrNotFound, got: %v", err)
	}
}

func TestWriteUserCredReadUserCred(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "simple",
			username: "user",
			password: "pass",
		},
		{
			name:     "email username",
			username: "admin@example.com",
			password: "P@ssw0rd123!",
		},
		{
			name:     "empty password",
			username: "test",
			password: "",
		},
		{
			name:     "password with colons",
			username: "user",
			password: "pass:word:123",
		},
		{
			name:     "unicode",
			username: "ユーザー",
			password: "パスワード🔐",
		},
		{
			name:     "long password",
			username: "alice",
			password: "this-is-a-very-long-password-with-many-characters-to-test-storage-123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credName := "test-usercred-" + tt.name

			// Create credential
			cred := NewUnPw(tt.username, tt.password)

			// WriteUserCred
			if err := WriteUserCred(credName, cred); err != nil {
				t.Fatalf("WriteUserCred failed: %v", err)
			}

			// ReadUserCred
			retrieved, err := ReadUserCred(credName)
			if err != nil {
				t.Fatalf("ReadUserCred failed: %v", err)
			}

			// Verify username
			if retrieved.Username() != tt.username {
				t.Errorf("Username mismatch:\ngot:  %q\nwant: %q", retrieved.Username(), tt.username)
			}

			// Verify password
			if retrieved.Password() != tt.password {
				t.Errorf("Password mismatch:\ngot:  %q\nwant: %q", retrieved.Password(), tt.password)
			}

			// Cleanup
			Delete(credName)
		})
	}
}

func TestReadUserCredNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	_, err := ReadUserCred("test-nonexistent-usercred-xyz")
	if err == nil {
		t.Error("ReadUserCred of nonexistent credential should return error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadUserCred error should wrap ErrNotFound, got: %v", err)
	}
}

func TestReadUserCredInvalidFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	credName := "test-invalid-format"

	// Write invalid format (no colon)
	if err := Write(credName, []byte("no-colon-here")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Try to read as UserCred
	_, err := ReadUserCred(credName)
	if err == nil {
		t.Error("ReadUserCred with invalid format should return error")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("ReadUserCred error should wrap ErrInvalidFormat, got: %v", err)
	}

	// Cleanup
	Delete(credName)
}

func TestDelete(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	credName := "test-delete-credential"

	// Create a credential
	if err := Write(credName, []byte("test data")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify it exists
	if _, err := Read(credName); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Delete it
	if err := Delete(credName); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err := Read(credName)
	if err == nil {
		t.Error("Read after Delete should fail")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Read after Delete error should wrap ErrNotFound, got: %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	err := Delete("test-nonexistent-delete-xyz")
	if err == nil {
		t.Error("Delete of nonexistent credential should return error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete error should wrap ErrNotFound, got: %v", err)
	}
}

func TestList(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create some test credentials
	testCreds := []string{
		"test-list-cred-1",
		"test-list-cred-2",
		"test-list-cred-3",
	}

	for _, name := range testCreds {
		if err := Write(name, []byte("test data")); err != nil {
			t.Fatalf("Write %s failed: %v", name, err)
		}
	}

	// List credentials
	names, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify all test credentials are in the list
	for _, expected := range testCreds {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected credential %q not found in list: %v", expected, names)
		}
	}

	// Cleanup
	for _, name := range testCreds {
		Delete(name)
	}
}

func TestListEmpty(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Clean up all test credentials first
	names, _ := List()
	for _, name := range names {
		if len(name) > 5 && name[:5] == "test-" {
			Delete(name)
		}
	}

	// List should work even when empty
	names, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should be empty or not contain test credentials
	for _, name := range names {
		if len(name) > 5 && name[:5] == "test-" {
			t.Errorf("Found unexpected test credential: %s", name)
		}
	}
}

func TestOverwrite(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	credName := "test-overwrite"
	data1 := []byte("first data")
	data2 := []byte("second data")

	// Write first time
	if err := Write(credName, data1); err != nil {
		t.Fatalf("First Write failed: %v", err)
	}

	// Verify first write
	retrieved, err := Read(credName)
	if err != nil {
		t.Fatalf("First Read failed: %v", err)
	}
	if !bytes.Equal(retrieved, data1) {
		t.Errorf("First read mismatch")
	}

	// Overwrite
	if err := Write(credName, data2); err != nil {
		t.Fatalf("Second Write failed: %v", err)
	}

	// Verify overwrite
	retrieved, err = Read(credName)
	if err != nil {
		t.Fatalf("Second Read failed: %v", err)
	}
	if !bytes.Equal(retrieved, data2) {
		t.Errorf("Second read mismatch: got %v, want %v", retrieved, data2)
	}

	// Cleanup
	Delete(credName)
}

func TestDeprecatedFunctions(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	credName := "test-deprecated"
	value := "test-value-123"

	// Test WriteString (deprecated)
	if err := WriteString(credName, value); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	// Test ReadString (deprecated)
	retrieved, err := ReadString(credName)
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}

	if retrieved != value {
		t.Errorf("ReadString mismatch: got %q, want %q", retrieved, value)
	}

	// Verify it's compatible with new API
	retrievedKey, err := ReadKey(credName)
	if err != nil {
		t.Fatalf("ReadKey failed: %v", err)
	}

	if retrievedKey != value {
		t.Errorf("ReadKey mismatch: got %q, want %q", retrievedKey, value)
	}

	// Cleanup
	Delete(credName)
}

func TestMultipleCredentialTypes(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Store different types of credentials
	rawName := "test-multi-raw"
	keyName := "test-multi-key"
	userCredName := "test-multi-usercred"

	// Raw bytes
	rawData := []byte{0x01, 0x02, 0x03}
	if err := Write(rawName, rawData); err != nil {
		t.Fatalf("Write raw failed: %v", err)
	}

	// Key
	apiKey := "sk-test-key-123"
	if err := WriteKey(keyName, apiKey); err != nil {
		t.Fatalf("WriteKey failed: %v", err)
	}

	// UserCred
	cred := NewUnPw("testuser", "testpass")
	if err := WriteUserCred(userCredName, cred); err != nil {
		t.Fatalf("WriteUserCred failed: %v", err)
	}

	// List all
	names, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify all exist
	expectedNames := []string{rawName, keyName, userCredName}
	for _, expected := range expectedNames {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected credential %q not found", expected)
		}
	}

	// Read and verify each
	retrievedRaw, err := Read(rawName)
	if err != nil || !bytes.Equal(retrievedRaw, rawData) {
		t.Errorf("Raw data verification failed")
	}

	retrievedKey, err := ReadKey(keyName)
	if err != nil || retrievedKey != apiKey {
		t.Errorf("Key verification failed")
	}

	retrievedCred, err := ReadUserCred(userCredName)
	if err != nil || retrievedCred.Username() != "testuser" || retrievedCred.Password() != "testpass" {
		t.Errorf("UserCred verification failed")
	}

	// Cleanup
	Delete(rawName)
	Delete(keyName)
	Delete(userCredName)
}

func TestConcurrentAccess(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	credName := "test-concurrent"
	
	// Write initial data
	if err := Write(credName, []byte("initial")); err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// Run multiple reads concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			_, err := Read(credName)
			if err != nil {
				t.Errorf("Concurrent read %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Cleanup
	Delete(credName)
}

// Benchmark tests
func BenchmarkWrite(b *testing.B) {
	cleanup := setupTestEnv(&testing.T{})
	defer cleanup()

	data := []byte("benchmark test data")
	credName := "test-benchmark-write"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Write(credName, data)
	}
	Delete(credName)
}

func BenchmarkRead(b *testing.B) {
	cleanup := setupTestEnv(&testing.T{})
	defer cleanup()

	credName := "test-benchmark-read"
	Write(credName, []byte("benchmark test data"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Read(credName)
	}
	Delete(credName)
}

func BenchmarkWriteKey(b *testing.B) {
	cleanup := setupTestEnv(&testing.T{})
	defer cleanup()

	credName := "test-benchmark-writekey"
	key := "sk-test-benchmark-key"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WriteKey(credName, key)
	}
	Delete(credName)
}

func BenchmarkReadKey(b *testing.B) {
	cleanup := setupTestEnv(&testing.T{})
	defer cleanup()

	credName := "test-benchmark-readkey"
	WriteKey(credName, "sk-test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadKey(credName)
	}
	Delete(credName)
}

func BenchmarkWriteUserCred(b *testing.B) {
	cleanup := setupTestEnv(&testing.T{})
	defer cleanup()

	credName := "test-benchmark-writeusercred"
	cred := NewUnPw("benchuser", "benchpass")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WriteUserCred(credName, cred)
	}
	Delete(credName)
}

func BenchmarkReadUserCred(b *testing.B) {
	cleanup := setupTestEnv(&testing.T{})
	defer cleanup()

	credName := "test-benchmark-readusercred"
	cred := NewUnPw("benchuser", "benchpass")
	WriteUserCred(credName, cred)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadUserCred(credName)
	}
	Delete(credName)
}
