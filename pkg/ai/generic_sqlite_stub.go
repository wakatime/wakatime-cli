//go:build (freebsd && !amd64 && !arm64) || (openbsd && !amd64 && !arm64) || netbsd || dragonfly

package ai

import "context"

func parseGenericAISQLite(context.Context, Parser, ParserConfig, []string) (Heartbeats, error) {
	return Heartbeats{}, nil
}
