//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import "context"

// Qoder contains params for detecting heartbeats from Qoder local activity.
type Qoder ParserConfig

// Parse is a no-op on BSD because the Qoder SQLite reader is not built there.
func (Qoder) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (Qoder) Name() string {
	return "Qoder"
}
