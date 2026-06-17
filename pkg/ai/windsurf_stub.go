//go:build netbsd || dragonfly || (freebsd && !(amd64 || arm64)) || (openbsd && !(amd64 || arm64))

package ai

import "context"

// Windsurf contains params for detecting heartbeats from Windsurf transcripts.
type Windsurf ParserConfig

// Parse is a no-op on unsupported BSD targets because the Windsurf SQLite reader is not built there.
func (Windsurf) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (Windsurf) Name() string {
	return "Windsurf"
}
