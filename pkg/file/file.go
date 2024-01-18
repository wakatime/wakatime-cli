//go:build !windows

package file

import "os"

// OpenNoLock opens a file for reading in non-exclusive mode. In Unix-like
// environments it just calls os.Open, but on Windows it forks syscall.Open
// for control over the sharemode, adding syscall.FILE_SHARE_DELETE.
func OpenNoLock(path string) (*os.File, error) {
	return os.Open(path) // nolint:gosec
}
