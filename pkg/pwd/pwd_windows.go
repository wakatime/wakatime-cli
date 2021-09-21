//go:build windows

package pwd

// Passwd represents an entry of the user database defined in <pwd.h>.
type Passwd struct {
	Name   string // user name
	Passwd string // user password
	UID    uint32 // user ID
	GID    uint32 // group ID
	Gecos  string // real name
	Dir    string // home directory
	Shell  string // shell program
}

// Getpwuid searches the user database for an entry with a matching uid.
func Getpwuid(uid uint32) *Passwd {
	return &Passwd{}
}
