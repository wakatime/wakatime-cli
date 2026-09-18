//go:build (freebsd && !amd64 && !arm64) || (openbsd && !amd64 && !arm64) || netbsd || dragonfly

package ai

import "context"

// ZCode contains parameters for detecting heartbeats from ZCode logs.
type ZCode ParserConfig

// Parse is a no-op on unsupported BSD targets because the ZCode SQLite reader is not built there.
func (ZCode) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns the ZCode parser name.
func (ZCode) Name() string { return "ZCode" }
