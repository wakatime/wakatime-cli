//go:build linux

package system

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"syscall"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// OSName returns the runtime machine's operating system name.
func OSName(ctx context.Context) string {
	os := runtime.GOOS

	var buf syscall.Utsname

	logger := log.Extract(ctx)

	err := syscall.Uname(&buf)
	if err != nil {
		logger.Debugf("Uname error: %s", err)

		return os
	}

	arr := buf.Sysname[:]
	output := make([]byte, 0, len(arr))

	for _, c := range arr {
		if c == 0x00 {
			break
		}

		output = append(output, byte(c))
	}

	alternateOS := string(output)
	if alternateOS != "" && !strings.EqualFold(alternateOS, os) {
		return fmt.Sprintf("%s-%s", alternateOS, os)
	}

	return os
}
