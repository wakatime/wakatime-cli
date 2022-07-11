//go:build !linux

package system

import (
	"runtime"
)

// OSName returns the runtime machine's operating system name.
func OSName() string {
	return runtime.GOOS
}
