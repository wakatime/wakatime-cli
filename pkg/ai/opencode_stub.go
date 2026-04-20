//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import "context"

// OpenCode contains params for detecting heartbeats from OpenCode session logs.
type OpenCode ParserConfig

// Parse is a no-op on BSD because the OpenCode SQLite reader is not built there.
func (OpenCode) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (OpenCode) Name() string {
	return "OpenCode"
}
