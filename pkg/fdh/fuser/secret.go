package fuser

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
)

const credentialName = "fuser-large-random-secret"

func getUserSecret() (string, error) {
	// Try to read existing credential
	secret, err := credmgr.ReadString(credentialName)
	if err == nil {
		return secret, nil
	}

	// Create new credential
	randomBytes := make([]byte, 128)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	secret = hex.EncodeToString(randomBytes)
	if err := credmgr.WriteString(credentialName, secret); err != nil {
		return "", err
	}
	return secret, nil
}
