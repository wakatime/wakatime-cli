package filestats

import (
	"context"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect filestats. At the
// moment only the total number of lines in a file is detected.
func WithDetection() heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute filestats detection")

			for n, h := range hh {
				if h.EntityType != heartbeat.FileType {
					continue
				}

				if h.IsUnsavedEntity {
					continue
				}

				if h.Lines != nil {
					continue
				}

				if h.IsRemote() {
					continue
				}

				filepath := h.Entity
				if h.LocalFile != "" {
					filepath = h.LocalFile
				}

				fileInfo, err := os.Stat(filepath)
				if err != nil {
					logger.Warnf("failed to retrieve file stats of file %q: %s", filepath, err)
					continue
				}

				if fileInfo.Size() > file.MaxFileSizeSupported {
					logger.Debugf(
						"file %q exceeds max file size of %d bytes. Lines won't be counted",
						h.Entity,
						file.MaxFileSizeSupported,
					)

					continue
				}

				lines, err := file.CountLines(ctx, filepath)
				if err != nil {
					logger.Warnf("failed to detect the total number of lines in file %q: %s", filepath, err)
					continue
				}

				hh[n].Lines = heartbeat.PointerTo(lines)
			}

			return next(ctx, hh)
		}
	}
}
