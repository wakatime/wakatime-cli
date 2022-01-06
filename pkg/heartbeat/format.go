package heartbeat

import (
	"path/filepath"
	"runtime"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/yookoala/realpath"
)

// WithFormatting initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to format entity's filepath.
func WithFormatting() HandleOption {
	return func(next Handle) Handle {
		return func(hh []Heartbeat) ([]Result, error) {
			log.Debugln("execute heartbeat filepath formatting")

			for n, h := range hh {
				hh[n] = Format(h)
			}

			return next(hh)
		}
	}
}

// Format accepts a heartbeat formats it's filepath and returns the formatted version.
func Format(h Heartbeat) Heartbeat {
	if h.EntityType == FileType && !windows.IsWindowsNetworkMount(h.Entity) {
		formatted, err := filepath.Abs(h.Entity)
		if err != nil {
			log.Warnf("failed to resolve absolute path for %q: %s", h.Entity, err)
		} else {
			h.Entity = formatted
		}

		// evaluate any symlinks
		formatted, err = realpath.Realpath(h.Entity)
		if err != nil {
			log.Warnf("failed to resolve real path for %q: %s", h.Entity, err)
		} else {
			h.Entity = formatted
		}
	}

	if h.EntityType == FileType && runtime.GOOS == "windows" {
		formatted, err := windows.FormatFilePath(h.Entity)
		if err != nil {
			log.Warnf("failed to format windows file path: %q: %s", h.Entity, err)
		} else {
			h.Entity = formatted
		}
	}

	if h.EntityType == FileType && runtime.GOOS == "windows" && !windows.IsWindowsNetworkMount(h.Entity) {
		var err error

		h.LocalFile, err = windows.FormatLocalFilePath(h.LocalFile, h.Entity)
		if err != nil {
			log.Warnf("failed to format local file path: %s", err)
		}
	}

	if h.EntityType == FileType && runtime.GOOS == "windows" && h.ProjectPathOverride != "" {
		formatted, err := windows.FormatFilePath(h.ProjectPathOverride)
		if err != nil {
			log.Warnf("failed to format windows file path: %q: %s", h.ProjectPathOverride, err)
		} else {
			h.ProjectPathOverride = formatted
		}
	}

	return h
}
