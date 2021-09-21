//go:build !windows

// Package pwd wraps system password functions `getpwnam(3)` and
// `getpwuid()`.  It was inspired by <https://github.com/Maki-Daisuke/go-pwd>.
package pwd

/*
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"
)

// Passwd is the Go type that corresponds to the C `struct passwd` defined in
// `pwd.h`; see man page `getpwnam(3)`.
type Passwd struct {
	Name   string
	Passwd string
	UID    uint32
	GID    uint32
	Gecos  string
	Dir    string
	Shell  string
}

func newPasswdFromC(c *C.struct_passwd) *Passwd {
	if c == nil {
		return nil
	}

	return &Passwd{
		Name:   C.GoString(c.pw_name),
		Passwd: C.GoString(c.pw_passwd),
		UID:    uint32(c.pw_uid),
		GID:    uint32(c.pw_uid),
		Gecos:  C.GoString(c.pw_gecos),
		Dir:    C.GoString(c.pw_dir),
		Shell:  C.GoString(c.pw_shell),
	}
}

// `mu` serializes calls to C functions that return statically allocated data
// that is overwritten in the next call.  The mutex must be held locked until
// the data has been copied to Go variables.
// nolint: gochecknoglobals
var mu = sync.Mutex{}

// Getpwnam calls `getpwnam(3)`.
func Getpwnam(name string) *Passwd {
	cName := C.CString(name)

	defer C.free(unsafe.Pointer(cName))

	mu.Lock()

	defer mu.Unlock()

	return newPasswdFromC(C.getpwnam(cName))
}

// Getpwuid calls `getpwuid()`; see man page `getpwnam(3)`.
func Getpwuid(uid uint32) *Passwd {
	mu.Lock()

	defer mu.Unlock()

	return newPasswdFromC(C.getpwuid(C.uint(uid)))
}
