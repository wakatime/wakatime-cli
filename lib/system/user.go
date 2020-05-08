package system

import "os"

// GetHomeDirectory Get user's home directory
func GetHomeDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return home
}
