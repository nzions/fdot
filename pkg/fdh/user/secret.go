package user

import "github.com/nzions/fdot/pkg/fdh/credmgr"

const credentialName = "fdot-user-secret"

func getUserSecret() (string, error) {
	// Try to read existing credential
	secret, err := credmgr.ReadString(credentialName)
	if err == nil {
		return secret, nil
	}

	// If credential doesn't exist, create a default one
	defaultSecret := "fdh"
	if err := credmgr.WriteString(credentialName, defaultSecret); err != nil {
		return "", err
	}

	return defaultSecret, nil
}
