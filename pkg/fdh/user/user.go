package user

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/nzions/fdot/pkg/fdh"
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
	dataDir := filepath.Join(homeDir, fdh.FDOTDir)
	if err := fdh.CheckCreateDir(dataDir); err != nil {
		panicMsg("dataDir", err)
	}

	// get network directory
	networkDir := filepath.Join(dataDir, "netcfg")
	if err := fdh.CheckCreateDir(networkDir); err != nil {
		panicMsg("networkDir", err)
	}

	secret, err := getUserSecret()
	if err != nil {
		panicMsg("getUserSecret", err)
	}

	CurrentUser = &FUser{
		Username:   user.Username,
		HomeDir:    homeDir,
		NetworkDir: networkDir,
		Secret:     secret,
	}
}

type FUser struct {
	Username   string
	HomeDir    string
	DataDir    string
	NetworkDir string
	Secret     string
}
