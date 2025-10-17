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

	CurrentUser = &FUser{
		Username:   user.Username,
		HomeDir:    homeDir,
		DataDir:    dataDir,
		NetworkDir: networkDir,
	}
}

type FUser struct {
	Username   string
	HomeDir    string
	DataDir    string
	NetworkDir string
}

func (u *FUser) BigKey() (string, error) {
	bigKey, err := credmgr.ReadKey(fdotconfig.BigKeySecretName)
	if err == nil {
		return bigKey, nil
	}

	// Create new big key
	randomBytes := make([]byte, 128)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	bigKey = hex.EncodeToString(randomBytes)
	if err := credmgr.WriteKey(fdotconfig.BigKeySecretName, bigKey); err != nil {
		return "", err
	}
	return bigKey, nil
}

func (u *FUser) SSHCreds() (username, password string, err error) {
	cred, err := credmgr.ReadUserCred(fdotconfig.SSHCredSecretName)
	if err != nil {
		return "", "", err
	}
	return cred.Username(), cred.Password(), nil
}

func (u *FUser) SetSSHCreds(username, password string) error {
	cred := credmgr.NewUnPw(username, password)
	return credmgr.WriteUserCred(fdotconfig.SSHCredSecretName, cred)
}

// CredFilePath returns the path to the encrypted credentials file
func (u *FUser) CredFilePath() string {
	return filepath.Join(u.DataDir, "credentials.enc")
}
