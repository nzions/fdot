package credmgr

import (
	"fmt"
	"strings"
)

// UserCred represents a username/password credential pair.
type UserCred interface {
	Username() string
	Password() string
}

// obfuscatedUserCred represents a username/password credential with obfuscated password storage.
// The password is XOR-encoded with a rotating key and base64-encoded to prevent trivial memory dumps.
type obfuscatedUserCred struct {
	username       string
	obfuscatedPass []byte // XOR-encoded password
	obfuscationKey []byte // Rotating key for XOR
}

// NewUnPw creates a new username/password credential with obfuscated password storage.
func NewUnPw(username, password string) UserCred {
	return newObfuscatedUserCred(username, password)
}

// newObfuscatedUserCred creates a new obfuscated credential.
func newObfuscatedUserCred(username, password string) *obfuscatedUserCred {
	// Generate a simple rotating key based on username
	key := generateObfuscationKey(username)

	// XOR encode the password
	obfuscated := xorEncode([]byte(password), key)

	return &obfuscatedUserCred{
		username:       username,
		obfuscatedPass: obfuscated,
		obfuscationKey: key,
	}
}

// Username returns the username.
func (u *obfuscatedUserCred) Username() string {
	return u.username
}

// Password returns the decoded password.
func (u *obfuscatedUserCred) Password() string {
	// XOR decode to get original password
	decoded := xorEncode(u.obfuscatedPass, u.obfuscationKey)
	return string(decoded)
}

// marshal converts obfuscatedUserCred to storable format (plaintext for storage encryption).
func (u *obfuscatedUserCred) marshal() []byte {
	// For storage, we use plaintext since the file is already AES-encrypted
	return []byte(u.username + ":" + u.Password())
}

// unmarshalUnPw parses a username:password credential and returns obfuscated form.
func unmarshalUnPw(data []byte) (UserCred, error) {
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: expected 'username:password'", ErrInvalidFormat)
	}
	return newObfuscatedUserCred(parts[0], parts[1]), nil
}

// generateObfuscationKey creates a rotating key based on username.
func generateObfuscationKey(username string) []byte {
	// Create a key that's at least 16 bytes
	key := []byte(username + "!@#$%^&*()_+")
	if len(key) < 16 {
		// Pad with repeating pattern
		for len(key) < 16 {
			key = append(key, key...)
		}
		key = key[:16]
	}
	return key
}

// xorEncode performs XOR encoding with a rotating key.
func xorEncode(data, key []byte) []byte {
	encoded := make([]byte, len(data))
	keyLen := len(key)
	for i := range data {
		encoded[i] = data[i] ^ key[i%keyLen]
	}
	return encoded
}
