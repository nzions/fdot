package credmgr

import (
	"fmt"
	"strings"
)

// UnPw represents a username/password credential pair.
type UnPw struct {
	username string
	password string
}

// NewUnPw creates a new username/password credential.
func NewUnPw(username, password string) *UnPw {
	return &UnPw{
		username: username,
		password: password,
	}
}

// Username returns the username.
func (u *UnPw) Username() string {
	return u.username
}

// Password returns the password.
func (u *UnPw) Password() string {
	return u.password
}

// marshal converts UnPw to storable format.
func (u *UnPw) marshal() []byte {
	return []byte(u.username + ":" + u.password)
}

// unmarshalUnPw parses a username:password credential.
func unmarshalUnPw(data []byte) (*UnPw, error) {
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: expected 'username:password'", ErrInvalidFormat)
	}
	return &UnPw{
		username: parts[0],
		password: parts[1],
	}, nil
}
