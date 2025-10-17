package fuser

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"os/user"
	"path/filepath"

	"github.com/nzions/fdot/pkg/fdh"
	"github.com/nzions/fdot/pkg/fdh/credmgr"
	"github.com/nzions/fdot/pkg/fdotconfig"
)

var CurrentUser *FUser

func panicMsg(msg string, err error) {
	panic("fdh.User.init " + msg + ": " + err.Error())
}

func init() {

	// get username
	user, err := user.Current()
	if err != nil {
		panicMsg("current user", err)
	}

	// get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panicMsg("homedir", err)
	}

	// get data directory
	dataDir := filepath.Join(homeDir, fdotconfig.FDOTDir)
	if err := fdh.CheckCreateDir(dataDir); err != nil {
		panicMsg("dataDir", err)
	}

	// get network directory
	networkDir := filepath.Join(dataDir, "netcfg")
	if err := fdh.CheckCreateDir(networkDir); err != nil {
		panicMsg("networkDir", err)
	}

	// create credential manager with custom path
	credFilePath := filepath.Join(dataDir, "credentials.enc")
	cm, err := credmgr.New(credFilePath)
	if err != nil {
		panicMsg("credmgr.New", err)
	}

	CurrentUser = &FUser{
		Username:    user.Username,
		HomeDir:     homeDir,
		DataDir:     dataDir,
		NetworkDir:  networkDir,
		CredManager: cm,
	}

	// Register as the path provider for credential operations
	fdotconfig.SetPathProvider(CurrentUser)
}

type FUser struct {
	Username    string
	HomeDir     string
	DataDir     string
	NetworkDir  string
	CredManager credmgr.CredManager // OO credential manager instance
}

func (u *FUser) BigKey() (string, error) {
	bigKey, err := u.CredManager.ReadKey(fdotconfig.BigKeySecretName)
	if err == nil {
		return bigKey, nil
	}

	// Create new big key
	randomBytes := make([]byte, 128)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	bigKey = hex.EncodeToString(randomBytes)
	if err := u.CredManager.WriteKey(fdotconfig.BigKeySecretName, bigKey); err != nil {
		return "", err
	}
	return bigKey, nil
}

func (u *FUser) SSHCreds() (credmgr.UserCred, error) {
	cred, err := u.CredManager.ReadUserCred(fdotconfig.SSHCredSecretName)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func (u *FUser) SetSSHCreds(username, password string) error {
	cred := credmgr.NewUnPw(username, password)
	return u.CredManager.WriteUserCred(fdotconfig.SSHCredSecretName, cred)
}

// CredFilePath returns the path to the encrypted credentials file
func (u *FUser) CredFilePath() string {
	return filepath.Join(u.DataDir, "credentials.enc")
}
