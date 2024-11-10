//go:build !linux

package system

import (
	"context"
	"runtime"
)

// OSName returns the runtime machine's operating system name.
func OSName(_ context.Context) string {
	return runtime.GOOS
}
