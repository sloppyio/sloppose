package homedir

import (
	"os"
	"os/user"
	"runtime"
)

// Get returns the home directory of the current user with the help of
// environment variables depending on the target operating system.
func Get() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	}

	homeDir := os.Getenv(env)
	if homeDir == "" && runtime.GOOS != "windows" {
		if u, err := user.Current(); err == nil {
			return u.HomeDir
		}
	}

	return homeDir
}
