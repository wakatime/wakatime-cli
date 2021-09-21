//go:build !windows

/*
Package pwd is a thin wrapper of C library <pwd.h>.
This is designed as thin as possible, but aimed to be thread-safe.
*/
package pwd

/*
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>

// While getpwuid requires "uid_t" according to man page, it actually requires
// "__uit_t" in the source code, that causes cgo compile error ("uid_t" is
// actually aliased to __uid_t).
// Unlike Linux, getpwuid on Mac OS X requires uid_t properly. For compatibility,
// we use a C function as a bridge here.
struct passwd *getpwuid_aux(unsigned int uid) {
	return getpwuid((uid_t)uid);
}
*/
import "C"

// Passwd represents an entry of the user database defined in <pwd.h>
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
	cpw := C.getpwuid_aux(C.uint(uid))
	if cpw != nil {
		return &Passwd{
			Name:   C.GoString(cpw.pw_name),
			Passwd: C.GoString(cpw.pw_passwd),
			UID:    uint32(cpw.pw_uid),
			GID:    uint32(cpw.pw_uid),
			Gecos:  C.GoString(cpw.pw_gecos),
			Dir:    C.GoString(cpw.pw_dir),
			Shell:  C.GoString(cpw.pw_shell),
		}
	}

	return nil
}
