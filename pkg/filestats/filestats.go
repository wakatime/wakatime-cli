package filestats

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/file"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Max file size supporting line number count stats. Files larger than this in
// bytes will not have a line count stat for performance. Default is 2MB (2*1024*1024).
const maxFileSizeSupported = 2097152

// WithDetection initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect filestats. At the
// moment only the total number of lines in a file is detected.
func WithDetection() heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute filestats detection")

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
					log.Warnf("failed to retrieve file stats of file %q: %s", filepath, err)
					continue
				}

				if fileInfo.Size() > maxFileSizeSupported {
					log.Debugf(
						"file %q exceeds max file size of %d bytes. Lines won't be counted",
						h.Entity,
						maxFileSizeSupported,
					)

					continue
				}

				lines, err := countLineNumbers(filepath)
				if err != nil {
					log.Warnf("failed to detect the total number of lines in file %q: %s", filepath, err)
					continue
				}

				hh[n].Lines = heartbeat.PointerTo(lines)
			}

			return next(hh)
		}
	}
}

func countLineNumbers(filepath string) (int, error) {
	f, err := file.OpenNoLock(filepath) // nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %s", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Debugf("failed to close file: %s", err)
		}
	}()

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
