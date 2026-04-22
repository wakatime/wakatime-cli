//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import "context"

// Windsurf contains params for detecting heartbeats from Windsurf transcripts.
type Windsurf ParserConfig

// Parse is a no-op on BSD because the Windsurf SQLite reader is not built there.
func (Windsurf) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (Windsurf) Name() string {
	return "Windsurf"
}
