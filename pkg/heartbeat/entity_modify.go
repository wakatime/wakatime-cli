package heartbeat

import (
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// WithEntityModifer initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to change an entity path.
func WithEntityModifer() HandleOption {
	return func(next Handle) Handle {
		return func(hh []Heartbeat) ([]Result, error) {
			log.Debugln("execute heartbeat entity modifier")

			for n, h := range hh {
				// Support XCode playgrounds
				if h.EntityType == FileType && isXCodePlayground(h.Entity) {
					hh[n].Entity = filepath.Join(h.Entity, "Contents.swift")
				}
			}

			return next(hh)
		}
	}
}

func isXCodePlayground(fp string) bool {
	if !(strings.HasSuffix(fp, ".playground") ||
		strings.HasSuffix(fp, ".xcplayground") ||
		strings.HasSuffix(fp, ".xcplaygroundpage")) {
		return false
	}

	return isDir(fp)
}
