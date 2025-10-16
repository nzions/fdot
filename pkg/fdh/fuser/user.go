package fuser

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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
	bigKey, err := credmgr.ReadString(fdotconfig.BigKeySecretName)
	if err == nil {
		return bigKey, nil
	}

	// Create new big key
	randomBytes := make([]byte, 128)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	bigKey = hex.EncodeToString(randomBytes)
	if err := credmgr.WriteString(fdotconfig.BigKeySecretName, bigKey); err != nil {
		return "", err
	}
	return bigKey, nil
}

func (u *FUser) SSHCreds() (username, password string, err error) {
	c, err := credmgr.Read(fdotconfig.SSHCredSecretName)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(string(c), ":")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid credential format: expected 'username:password'")
	}

	un := parts[0]
	pw := strings.Join(parts[1:], ":")
	return un, pw, nil
}

func (u *FUser) SetSSHCreds(username, password string) error {
	// Store credentials in the format: "username:password"
	sshCreds := username + ":" + password
	return credmgr.WriteString(fdotconfig.SSHCredSecretName, sshCreds)
}
