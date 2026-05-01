//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import "context"

// Cody contains params for detecting heartbeats from Cody by Sourcegraph transcripts.
type Cody ParserConfig

// Parse is a no-op on BSD because the Cody SQLite reader is not built there.
func (Cody) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its name.
func (Cody) Name() string {
	return "Cody"
}
