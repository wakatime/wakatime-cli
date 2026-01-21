package heartbeat

import (
	"context"
	"path/filepath"
	"strings"
)

// WithEntityModifier initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to change an entity path.
func WithEntityModifier() HandleOption {
	return func(next Handle) Handle {
		return func(ctx context.Context, hh []Heartbeat) ([]Result, error) {
			// logger := log.Extract(ctx)
			// logger.Debugln("execute heartbeat entity modifier")
			for n, h := range hh {
				// Support XCode playgrounds
				if h.EntityType == FileType && isXCodePlayground(ctx, h.Entity) {
					hh[n].Entity = filepath.Join(h.Entity, "Contents.swift")
				}

				// Support XCode projects
				if h.EntityType == FileType && isXCodeProject(ctx, h.Entity) {
					hh[n].Entity = filepath.Join(h.Entity, "project.pbxproj")
				}
			}

			return next(ctx, hh)
		}
	}
}

func isXCodePlayground(ctx context.Context, fp string) bool {
	if !strings.HasSuffix(strings.ToLower(fp), ".playground") &&
		!strings.HasSuffix(strings.ToLower(fp), ".xcplayground") &&
		!strings.HasSuffix(strings.ToLower(fp), ".xcplaygroundpage") {
		return false
	}

	return isDir(ctx, fp)
}

func isXCodeProject(ctx context.Context, fp string) bool {
	if !(strings.HasSuffix(strings.ToLower(fp), ".xcodeproj")) {
		return false
	}

	return isDir(ctx, fp)
}
