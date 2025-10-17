package credmgr

import (
	"bytes"
	"testing"
)

func TestNewUnPw(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "simple credentials",
			username: "john.doe",
			password: "secretpass123",
		},
		{
			name:     "special characters in password",
			username: "admin",
			password: "P@ssw0rd!#$%^&*()",
		},
		{
			name:     "unicode in password",
			username: "user",
			password: "パスワード🔐",
		},
		{
			name:     "empty password",
			username: "testuser",
			password: "",
		},
		{
			name:     "long password",
			username: "alice",
			password: "this-is-a-very-long-password-with-many-characters-to-test-edge-cases-123456789",
		},
		{
			name:     "short username",
			username: "a",
			password: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := NewUnPw(tt.username, tt.password)
			if cred == nil {
				t.Fatal("NewUnPw returned nil")
			}

			// Verify username
			if got := cred.Username(); got != tt.username {
				t.Errorf("Username() = %q, want %q", got, tt.username)
			}

			// Verify password
			if got := cred.Password(); got != tt.password {
				t.Errorf("Password() = %q, want %q", got, tt.password)
			}
		})
	}
}

func TestPasswordObfuscation(t *testing.T) {
	username := "testuser"
	password := "mySecretPassword123"

	cred := newObfuscatedUserCred(username, password)

	// Password should be obfuscated in memory (XOR-encoded)
	// The obfuscatedPass field should NOT contain the plaintext password
	if bytes.Contains(cred.obfuscatedPass, []byte(password)) {
		t.Error("Password is not obfuscated - plaintext found in obfuscatedPass field")
	}

	// But Password() method should return the original
	if got := cred.Password(); got != password {
		t.Errorf("Password() = %q, want %q", got, password)
	}

	// Verify obfuscation is reversible
	decoded := xorEncode(cred.obfuscatedPass, cred.obfuscationKey)
	if string(decoded) != password {
		t.Errorf("XOR decode failed: got %q, want %q", string(decoded), password)
	}
}

func TestPasswordObfuscationUniqueness(t *testing.T) {
	// Different usernames should produce different obfuscation for same password
	password := "samePassword123"
	
	cred1 := newObfuscatedUserCred("user1", password)
	cred2 := newObfuscatedUserCred("user2", password)

	// Obfuscated passwords should be different (different keys)
	if bytes.Equal(cred1.obfuscatedPass, cred2.obfuscatedPass) {
		t.Error("Same password with different usernames produced identical obfuscation")
	}

	// But both should decode to the same password
	if cred1.Password() != password || cred2.Password() != password {
		t.Error("Obfuscated passwords don't decode correctly")
	}
}

func TestGenerateObfuscationKey(t *testing.T) {
	tests := []struct {
		name     string
		username string
		minLen   int
	}{
		{"short username", "a", 16},
		{"normal username", "john.doe", 16},
		{"long username", "verylongusernamewithtonsofcharacters", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generateObfuscationKey(tt.username)
			
			if len(key) < tt.minLen {
				t.Errorf("generateObfuscationKey(%q) length = %d, want >= %d", 
					tt.username, len(key), tt.minLen)
			}

			// Key should be deterministic
			key2 := generateObfuscationKey(tt.username)
			if !bytes.Equal(key, key2) {
				t.Error("generateObfuscationKey is not deterministic")
			}

			// Different usernames should produce different keys
			if tt.username != "a" {
				differentKey := generateObfuscationKey("different")
				if bytes.Equal(key, differentKey) {
					t.Error("Different usernames produced same obfuscation key")
				}
			}
		})
	}
}

func TestXorEncode(t *testing.T) {
	tests := []struct {
		name string
		data string
		key  string
	}{
		{"simple", "hello", "key123"},
		{"empty data", "", "key"},
		{"data longer than key", "this is a long message", "short"},
		{"unicode", "こんにちは", "key"},
		{"special chars", "!@#$%^&*()", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(tt.data)
			key := []byte(tt.key)

			// Encode
			encoded := xorEncode(data, key)

			// Encoded should be different from original (unless data is empty)
			if len(data) > 0 && bytes.Equal(data, encoded) {
				t.Error("XOR encoding produced same output as input")
			}

			// Decode (XOR is symmetric)
			decoded := xorEncode(encoded, key)

			// Should get back original
			if !bytes.Equal(data, decoded) {
				t.Errorf("XOR decode failed: got %q, want %q", decoded, data)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "simple",
			username: "user",
			password: "pass",
			expected: "user:pass",
		},
		{
			name:     "with special chars",
			username: "admin@example.com",
			password: "P@ssw0rd!",
			expected: "admin@example.com:P@ssw0rd!",
		},
		{
			name:     "empty password",
			username: "test",
			password: "",
			expected: "test:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := newObfuscatedUserCred(tt.username, tt.password)
			marshaled := cred.marshal()

			if string(marshaled) != tt.expected {
				t.Errorf("marshal() = %q, want %q", marshaled, tt.expected)
			}
		})
	}
}

func TestUnmarshalUnPw(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		wantUser    string
		wantPass    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid simple",
			data:     "user:pass",
			wantUser: "user",
			wantPass: "pass",
			wantErr:  false,
		},
		{
			name:     "valid with colon in password",
			data:     "admin:pass:word:123",
			wantUser: "admin",
			wantPass: "pass:word:123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			data:     "user:",
			wantUser: "user",
			wantPass: "",
			wantErr:  false,
		},
		{
			name:        "missing colon",
			data:        "userpass",
			wantErr:     true,
			errContains: "invalid credential format",
		},
		{
			name:        "empty string",
			data:        "",
			wantErr:     true,
			errContains: "invalid credential format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := unmarshalUnPw([]byte(tt.data))

			if tt.wantErr {
				if err == nil {
					t.Error("unmarshalUnPw() expected error, got nil")
				} else if tt.errContains != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errContains)) {
					t.Errorf("unmarshalUnPw() error = %q, want containing %q", 
						err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unmarshalUnPw() unexpected error: %v", err)
			}

			if cred.Username() != tt.wantUser {
				t.Errorf("Username() = %q, want %q", cred.Username(), tt.wantUser)
			}

			if cred.Password() != tt.wantPass {
				t.Errorf("Password() = %q, want %q", cred.Password(), tt.wantPass)
			}
		})
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	tests := []struct {
		username string
		password string
	}{
		{"user", "pass"},
		{"admin@example.com", "complex!P@ssw0rd#123"},
		{"test", ""},
		{"a", "b"},
		{"user", "pass:with:colons"},
		{"unicode", "パスワード🔐"},
	}

	for _, tt := range tests {
		t.Run(tt.username+":"+tt.password, func(t *testing.T) {
			// Create credential
			cred := newObfuscatedUserCred(tt.username, tt.password)

			// Marshal
			marshaled := cred.marshal()

			// Unmarshal
			restored, err := unmarshalUnPw(marshaled)
			if err != nil {
				t.Fatalf("unmarshalUnPw() error: %v", err)
			}

			// Verify
			if restored.Username() != tt.username {
				t.Errorf("Username() = %q, want %q", restored.Username(), tt.username)
			}
			if restored.Password() != tt.password {
				t.Errorf("Password() = %q, want %q", restored.Password(), tt.password)
			}
		})
	}
}

func TestUserCredInterface(t *testing.T) {
	// Verify that obfuscatedUserCred implements UserCred interface
	var _ UserCred = (*obfuscatedUserCred)(nil)

	// Verify NewUnPw returns UserCred interface
	var cred UserCred = NewUnPw("user", "pass")
	if cred == nil {
		t.Fatal("NewUnPw returned nil")
	}

	// Interface methods should work
	if cred.Username() != "user" {
		t.Errorf("Username() = %q, want %q", cred.Username(), "user")
	}
	if cred.Password() != "pass" {
		t.Errorf("Password() = %q, want %q", cred.Password(), "pass")
	}
}

// Benchmark tests
func BenchmarkNewUnPw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUnPw("testuser", "testpassword123")
	}
}

func BenchmarkPasswordDecode(b *testing.B) {
	cred := newObfuscatedUserCred("user", "password123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cred.Password()
	}
}

func BenchmarkXorEncode(b *testing.B) {
	data := []byte("this is a test password with some length to it")
	key := []byte("obfuscationkey123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		xorEncode(data, key)
	}
}

func BenchmarkMarshalUnmarshal(b *testing.B) {
	cred := newObfuscatedUserCred("testuser", "testpassword123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marshaled := cred.marshal()
		_, _ = unmarshalUnPw(marshaled)
	}
}
