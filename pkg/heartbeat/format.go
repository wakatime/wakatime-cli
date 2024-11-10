package heartbeat

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
)

// WithFormatting initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to format entity's filepath.
func WithFormatting() HandleOption {
	return func(next Handle) Handle {
		return func(ctx context.Context, hh []Heartbeat) ([]Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute heartbeat filepath formatting")

			for n, h := range hh {
				if h.EntityType != FileType {
					continue
				}

				if h.IsRemote() {
					continue
				}

				hh[n] = Format(ctx, h)
			}

			return next(ctx, hh)
		}
	}
}

// Format accepts a heartbeat formats it's filepath and returns the formatted version.
func Format(ctx context.Context, h Heartbeat) Heartbeat {
	if !h.IsUnsavedEntity && (runtime.GOOS != "windows" || !windows.IsWindowsNetworkMount(h.Entity)) {
		formatLinuxFilePath(ctx, &h)
	}

	if runtime.GOOS == "windows" {
		formatWindowsFilePath(ctx, &h)
	}

	return h
}

func formatLinuxFilePath(ctx context.Context, h *Heartbeat) {
	logger := log.Extract(ctx)

	formatted, err := filepath.Abs(h.Entity)
	if err != nil {
		logger.Debugf("failed to resolve absolute path for %q: %s", h.Entity, err)
		return
	}

	h.Entity = formatted

	// evaluate any symlinks
	formatted, err = realpath.Realpath(h.Entity)
	if err != nil {
		logger.Debugf("failed to resolve real path for %q: %s", h.Entity, err)
		return
	}

	h.Entity = formatted

	if h.ProjectPathOverride != "" {
		formatted, err = filepath.Abs(h.ProjectPathOverride)
		if err != nil {
			logger.Debugf("failed to resolve absolute path for %q: %s", h.ProjectPathOverride, err)
			return
		}

		h.ProjectPathOverride = formatted

		// evaluate any symlinks
		formatted, err = realpath.Realpath(h.ProjectPathOverride)
		if err != nil {
			logger.Debugf("failed to resolve real path for %q: %s", h.ProjectPathOverride, err)
			return
		}

		h.ProjectPathOverride = formatted
	}
}

func formatWindowsFilePath(ctx context.Context, h *Heartbeat) {
	logger := log.Extract(ctx)

	h.Entity = windows.FormatFilePath(h.Entity)

	if !h.IsUnsavedEntity && !windows.IsWindowsNetworkMount(h.Entity) {
		var err error

		h.LocalFile, err = windows.FormatLocalFilePath(h.LocalFile, h.Entity)
		if err != nil {
			logger.Debugf("failed to format local file path: %s", err)
		}
	}

	if h.ProjectPathOverride != "" {
		h.ProjectPathOverride = windows.FormatFilePath(h.ProjectPathOverride)
	}
}
